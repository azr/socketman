package socketman_test

import (
	"crypto/tls"
	"io"
	"testing"

	"github.com/azr/socketman"
	"github.com/azr/socketman/internal"
)

var (
	aespool socketman.CypherPool
)

func init() {
	var err error
	aespool, err = socketman.NewAESPool([]byte("example key 1234"))
	if err != nil {
		panic(err)
	}
}

func TestCypher(t *testing.T) {
	server := &socketman.Server{
		Config: socketman.Config{
			CypherPool: aespool,
		},
	}
	client := &socketman.Client{}

	{
		//this works because server sends
		//encrypted stuff to client that will
		//send it back as is, server wil just decrypt it
		testEchoClient(t, server, client)
		//vice versa
		testEchoServer(t, server, client)
	}

	// test that cliens doesn't understands
	// when cypher not configured.
	in := "hello, world!"
	out := make([]byte, len(in))

	test(t, server, func(c io.ReadWriter) {
		for i := 0; i < len(in); {
			w, err := io.WriteString(c, in)
			if err != nil {
				t.Errorf("write failed: %s", err)
			}
			i += w
		}
	}, client, func(c io.ReadWriter) {
		for i := 0; i < len(in); {
			r, err := c.Read(out)
			if err != nil {
				t.Errorf("read failed: %s", err)
			}
			i += r
		}
	})

	if string(out) == in {
		t.Fatalf("Expected weirdly encoded stuff, not %s", out)
	}

	// configure client so it understands now !
	client.Config.CypherPool = server.Config.CypherPool

	test(t, server, func(c io.ReadWriter) {
		for i := 0; i < len(in); {
			w, err := io.WriteString(c, in)
			if err != nil {
				t.Errorf("write failed: %s", err)
			}
			i += w
		}
	}, client, func(c io.ReadWriter) {
		for i := 0; i < len(in); {
			r, err := c.Read(out)
			if err != nil {
				t.Errorf("read failed: %s", err)
			}
			i += r
		}
	})

	if string(out) != in {
		t.Fatalf("failed decrypting stuff sent by server: expected :'%s', got '%s'", in, out)
	}
}

func TestListenAndServe_with_tls_and_cypher(t *testing.T) {
	cert, err := tls.X509KeyPair(internal.LocalhostCert, internal.LocalhostKey)
	if err != nil {
		t.Fatal(err)
	}
	server := &socketman.Server{
		Config: socketman.Config{
			TLSConfig: &tls.Config{
				Certificates: []tls.Certificate{cert},
			},
			CypherPool: aespool,
		},
	}
	client := &socketman.Client{
		Config: socketman.Config{
			TLSConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			CypherPool: aespool,
		},
	}

	testEchoServer(t, server, client)
}
