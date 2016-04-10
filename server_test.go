package socketman_test

import (
	"io"
	"testing"
	"time"

	"sync"

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

func TestListenAndServe_echo_server(t *testing.T) {
	addr := "127.0.0.1:1234"

	//start echo server
	serverTasks := sync.WaitGroup{}
	serverTasks.Add(1)

	s := socketman.Server{}
	go func() {
		defer serverTasks.Done()
		if err := s.ListenAndServeFunc(addr, echoHandler); err != nil {
			t.Logf("ListenAndServeFunc returned: %s.", err)
		}
	}()

	time.Sleep(time.Millisecond) // sleep a little to be more sure server was started.

	in := "hello, world!"
	out := make([]byte, len(in))

	//start sending client
	//that will also check echo in return
	c := socketman.Client{}
	clientTasks := sync.WaitGroup{}
	clientTasks.Add(1)
	c.ConnectFunc(addr, func(c io.ReadWriter) {
		defer clientTasks.Done()
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
	})
	clientTasks.Wait()

	s.Close() // close server to see if it returns
	serverTasks.Wait()

	//check echo
	if string(out) != in {
		t.Fatalf("failed reading with simple echo handler: expected :%s, got %s", in, out)
	}
}

func TestListenAndServe_echo_client(t *testing.T) {
}

func TestListenAndServePanic(t *testing.T) {
	addr := "127.0.0.1:1235"
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
	addr := "127.0.0.1:1235"
	s := socketman.Server{
		Config: socketman.Config{
			IdleDeadline: time.Millisecond * 20,
		},
	}

	go func() {
		if err := s.ListenAndServeFunc(addr, echoHandler); err != nil {
			t.Fatalf("could not start server: %s. tests already running ?", err)
		}
	}()
	defer s.Close()
	time.Sleep(time.Millisecond)

	in := "hello, world!"
	out := make([]byte, len(in))

	//start sending client
	//that will also check echo in return
	c := socketman.Client{}
	clientTasks := sync.WaitGroup{}
	clientTasks.Add(1)
	var err error
	c.ConnectFunc(addr, func(c io.ReadWriter) {
		time.Sleep(s.Config.IdleDeadline)
		defer clientTasks.Done()
		for i := 0; i < len(in); {
			w := 0
			w, err = io.WriteString(c, in)
			if err != nil {
				break
			}
			i += w
		}

		for i := 0; i < len(in); {
			r := 0
			r, err = c.Read(out)
			if err != nil {
				break
			}
			i += r
		}
	})
	clientTasks.Wait()
	if err == nil {
		t.Errorf("read or write should have failed !")
	}

	//start sending client
	//that will also check echo in return
	clientTasks.Add(1)
	c.ConnectFunc(addr, func(c io.ReadWriter) {
		time.Sleep(s.Config.IdleDeadline / 2)
		defer clientTasks.Done()
		for i := 0; i < len(in); {
			w := 0
			w, err = io.WriteString(c, in)
			if err != nil {
				break
			}
			i += w
		}

		for i := 0; i < len(in); {
			r := 0
			r, err = c.Read(out)
			if err != nil {
				break
			}
			i += r
		}

		time.Sleep(s.Config.IdleDeadline / 2)
		for i := 0; i < len(in); {
			w := 0
			w, err = io.WriteString(c, in)
			if err != nil {
				break
			}
			i += w
		}

		for i := 0; i < len(in); {
			r := 0
			r, err = c.Read(out)
			if err != nil {
				break
			}
			i += r
		}
	})
	clientTasks.Wait()
	if err != nil && err != io.EOF {
		t.Errorf("active request should not have gone timeout. err: %s", err)
	}
}
