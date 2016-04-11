package socketman

import (
	"io"
	"log"
	"net"
	"time"
)

type conn struct {
	netCon net.Conn
	Config
}

func (c *conn) resetDeadline() {
	err := c.netCon.SetDeadline(time.Now().Add(c.Config.IdleDeadline))
	if err != nil {
		log.Printf("socketman: SetDeadline failed: %s", err)
	}
}

func (c *conn) Write(b []byte) (n int, err error) {
	if c.Config.IdleDeadline != 0 {
		defer func() {
			if err == nil {
				c.resetDeadline()
			}
		}()
	}
	return c.netCon.Write(b)
}
func (c *conn) Read(b []byte) (n int, err error) {
	if c.Config.IdleDeadline != 0 {
		defer func() {
			if err == nil {
				c.resetDeadline()
			}
		}()
	}
	return c.netCon.Read(b)
}

//A Handler handles socket comunications.
//Client & Server will close socket after the handler returns.
//
// For a server, if ServeSocket panics, the server (the caller of ServeSocket) assumes
// that the effect of the panic was isolated to the active socket.
// It recovers the panic, logs a stack trace to the server error log,
// and hangs up the connection.
// For a client, it's up to you.
type Handler interface {
	ServeSocket(io.ReadWriter)
}

// The HandlerFunc type is an adapter to allow the use of
// ordinary functions as Socket handlers.  If f is a function
// with the appropriate signature, HandlerFunc(f) is a
// Handler that calls f.
type HandlerFunc func(io.ReadWriter)

// ServeSocket calls f
func (f HandlerFunc) ServeSocket(c io.ReadWriter) {
	f(c)
}
