package socketman_test

import (
	"crypto/tls"
	"io"
	"testing"
	"time"

	"sync"

	"github.com/azr/socketman"
	"github.com/azr/socketman/internal"
)

func TestListenAndServe_echo(t *testing.T) {
	testEchoServer(t, &socketman.Server{}, &socketman.Client{})
	testEchoClient(t, &socketman.Server{}, &socketman.Client{})
}

func TestListenAndServePanic(t *testing.T) {
	addr := "127.0.0.1:1234"
	s := socketman.Server{}

	go func() {
		if err := s.ListenAndServeFunc(addr, panicHandler); err != nil {
			t.Fatalf("could not start server: %s. tests already running ?", err)
		}
	}()
	defer s.Close()
	time.Sleep(time.Millisecond)

	c := socketman.Client{}
	wg := sync.WaitGroup{}
	wg.Add(1)
	c.ConnectFunc(addr, func(c io.ReadWriter) {
		wg.Done()
	})
	wg.Wait()
	// all went fine !
	time.Sleep(time.Second)
}

func TestListenAndServeTimeout(t *testing.T) {
	server := &socketman.Server{
		Config: socketman.Config{
			IdleTimeout: 200 * time.Millisecond,
		},
	}

	in := "hello, world!"
	out := make([]byte, len(in))

	//start sending client
	//that will also check echo in return
	client := &socketman.Client{}

	var err error
	test(t, server, echoHandler, client, func(c io.ReadWriter) {
		time.Sleep(server.Config.IdleTimeout + 20*time.Millisecond)
		for i := 0; i < len(in) && err == nil; {
			w := 0
			w, err = io.WriteString(c, in)
			i += w
		}

		for i := 0; i < len(in) && err == nil; {
			r := 0
			r, err = c.Read(out)
			i += r
		}
	})
	if err == nil {
		t.Errorf("read or write should have failed !")
	}

	//start sending client
	//that will also check echo in return
	//client should not go idle
	err = nil
	test(t, server, echoHandler, client, func(c io.ReadWriter) {
		for j := 0; j < 2; j++ {
			time.Sleep(time.Duration(float64(server.Config.IdleTimeout) * 0.90)) // Read & Write should not wait more that IdleTimeout
			for i := 0; i < len(in) && err == nil; {
				w := 0
				w, err = io.WriteString(c, in)
				i += w
			}
			for i := 0; i < len(in) && err == nil; {
				r := 0
				r, err = c.Read(out)
				i += r
			}
		}
	})
	if err != nil && err != io.EOF {
		t.Errorf("active request should not have gone timeout. err: %s", err)
	}
	if in != string(out) {
		t.Fatalf("server failed echoing: expected :%s, got %s", in, out)
	}
}

func TestListenAndServe_with_tls(t *testing.T) {
	cert, err := tls.X509KeyPair(internal.LocalhostCert, internal.LocalhostKey)
	if err != nil {
		t.Fatal(err)
	}
	server := &socketman.Server{
		Config: socketman.Config{
			TLSConfig: &tls.Config{
				Certificates: []tls.Certificate{cert},
			},
		},
	}
	client := &socketman.Client{
		Config: socketman.Config{
			TLSConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	testEchoServer(t, server, client)
}
