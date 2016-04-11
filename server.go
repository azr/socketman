//Package socketman implements simple websocket client/server in golang
//
//And I think it's gonna be a long, long, time. ♪♫♬
package socketman

import (
	"crypto/tls"
	"io"
	"log"
	"net"
	"runtime"
	"time"

	"sync"

	"golang.org/x/net/context"
)

//Server is a socket server
type Server struct {
	//Config is a configuration for new incoming connections
	Config

	//Context represents the context of a server
	//this context is cancelled and reset upon Close.
	//
	//When nil, ListenAndServe will just set it with a context.Background()
	Context context.Context

	//cancelContext will be initialised on first
	//ListenAndServe call.
	cancelContext func()

	// mu guards Context and cancelContext
	mu sync.Mutex
}

//ListenAndServe listens on the TCP network address addr and
//then calls handler to handle requests on incoming connections.
//
//ListenAndServe blocks.
//
//ListenAndServe can be called multiple time on different addrs, one
//Close call will close them all.
//If configuration is changed between two ListenAndServe calls, already
//running servers will just keep running with old config. Changing
//configuration after a ListenAndServe call might cause races.
//
//Otherwise ListenAndServe is thread safe.
//
//The syntax of laddr is "host:port", like "127.0.0.1:8080".
//If host is omitted, as in ":8080", ListenAndServe listens on all available
//interfaces instead of just the interface with the given host address.
//See net.Dial for more details about address syntax.
func (s *Server) ListenAndServe(addr string, handler Handler) error {

	s.mu.Lock()
	if s.Context == nil {
		s.Context = context.Background()
	}

	if s.cancelContext == nil {
		s.Context, s.cancelContext = context.WithCancel(s.Context)
	}
	s.mu.Unlock()

	// listen using tcp because we need to make sure order
	// and integrity is kept. Thanks tcp !
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	if s.Config.TLSConfig != nil {
		config := cloneTLSConfig(s.Config.TLSConfig)
		tlsListener := tls.NewListener(tcpKeepAliveListener{listener.(*net.TCPListener)}, config)
		return s.Serve(tlsListener, handler)
	}
	return s.Serve(tcpKeepAliveListener{listener.(*net.TCPListener)}, handler)
}

// Serve accepts incoming connections on the Listener l, creating a
// new service goroutine for each. The service goroutines read requests and
// then call handler to reply to them.
// Serve always returns a non-nil error.
func (s *Server) Serve(l net.Listener, handler Handler) error {
	go func() {
		<-s.Context.Done()
		l.Close()
	}()
	var tempDelay time.Duration // how long to sleep on accept failure
	for {
		c, e := l.Accept()
		if e != nil {
			if ne, ok := e.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				log.Printf("socketman: Accept error: %v; retrying in %v", e, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return e
		}
		tempDelay = 0

		if s.Config.IdleDeadline != 0 {
			e = c.SetDeadline(time.Now().Add(s.Config.IdleDeadline))
			if e != nil {
				log.Printf("failed to set idle timeout: %s.", e)
			}
		}
		go func() {
			defer func() {
				if err := recover(); err != nil {
					const size = 64 << 10
					buf := make([]byte, size)
					buf = buf[:runtime.Stack(buf, false)]
					log.Printf("socketman: panic serving %v: %v\n%s", l.Addr(), err, buf)
				}
			}()
			handler.ServeSocket(c)
			err := c.Close()
			if err != nil {
				log.Printf("socketman: connection close failed: %s", err)
			}
		}()
	}
}

//ListenAndServeFunc callsListenAndServe with a plain func
func (s *Server) ListenAndServeFunc(addr string, handler func(io.ReadWriter)) error {
	return s.ListenAndServe(addr, HandlerFunc(handler))
}

// Close closes the server.
// server will stop listenning for new connections.
// any ongoing connection will keep running.
func (s *Server) Close() {
	if s.cancelContext != nil {
		s.cancelContext()
	}
	s.cancelContext = nil
	s.Context = nil
}
