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

var (
	echoHandler = func(c io.ReadWriter) {
		io.Copy(c, c)
	}
	panicHandler = func(c io.ReadWriter) {
		panic("what is the purpose of life ????")
	}
)

func test(t *testing.T, server *socketman.Server, serverHandler func(io.ReadWriter), client *socketman.Client, clientHandler func(io.ReadWriter)) {
	addr := "127.0.0.1:1234"

	//start server
	serverTasks := sync.WaitGroup{}
	serverTasks.Add(1)

	go func() {
		defer serverTasks.Done()
		if err := server.ListenAndServeFunc(addr, serverHandler); err != nil {
			t.Logf("ListenAndServeFunc returned: %s.", err)
		}
	}()

	time.Sleep(time.Millisecond) // sleep a little to be more sure server was started.

	clientTasks := sync.WaitGroup{}
	clientTasks.Add(1)
	err := client.ConnectFunc(addr, func(c io.ReadWriter) {
		defer clientTasks.Done()
		clientHandler(c)
	})
	if err != nil {
		t.Errorf("client.ConnectFunc failed: %s", err)
		clientTasks.Done()
	}

	clientTasks.Wait()
	server.Close()
	serverTasks.Wait()
}

func testEchoServer(t *testing.T, server *socketman.Server, client *socketman.Client) {
	in := "hello, world!"
	out := make([]byte, len(in))

	test(t, server, echoHandler, client, func(c io.ReadWriter) {
		for i := 0; i < len(in); {
			w, err := io.WriteString(c, in)
			if err != nil {
				t.Errorf("write failed: %s", err)
			}
			if w == 0 {
				break // nothing to do anymore !
			}
			i += w
		}

		for i := 0; i < len(in); {
			r, err := c.Read(out)
			if err != nil {
				t.Errorf("read failed: %s", err)
			}
			if r == 0 {
				break // nothing to do anymore !
			}
			i += r
		}
	})

	if string(out) != in {
		t.Fatalf("failed reading with simple echo handler: expected :%s, got %s", in, out)
	}
}

func testEchoClient(t *testing.T, server *socketman.Server, client *socketman.Client) {
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

		for i := 0; i < len(in); {
			r, err := c.Read(out)
			if err != nil {
				t.Errorf("read failed: %s", err)
			}
			i += r
		}
	}, client, echoHandler)

	if string(out) != in {
		t.Fatalf("failed reading with simple echo handler: expected :%s, got %s", in, out)
	}
}

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
			IdleTimeout: time.Second,
		},
	}

	in := "hello, world!"
	out := make([]byte, len(in))

	//start sending client
	//that will also check echo in return
	client := &socketman.Client{}

	var err error
	test(t, server, echoHandler, client, func(c io.ReadWriter) {
		time.Sleep(server.Config.IdleTimeout)
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
			time.Sleep(server.Config.IdleTimeout / 2) // Read & Write should not wait more that IdleTimeout
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
