package socketman_test

import (
	"io"
	"sync"
	"testing"
	"time"

	"github.com/azr/socketman"
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
	//TODO: return any error so we can check them.
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
		t.Logf("got stuff: %s", out)
	}, client, echoHandler)

	if string(out) != in {
		t.Fatalf("failed reading with simple echo handler: expected :%s, got %s", in, out)
	}
}
