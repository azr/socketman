package socketman

import (
	"io"
	"log"
	"net"
	"time"
)

//net con embeds a net.Conn
// it allows to bump I/O deadline
// after each successfull read/write.
// in client and/or server.
// if config containts a CypherPool
// one reader and one writer will be
// instantiated with pool and will embed
// the net.Conn. This allows encrypting
// sent messages.
type conn struct {
	netCon net.Conn
	w      io.Writer
	r      io.Reader
	io.Closer
	Config
}

func newconn(netConn net.Conn, conf Config) *conn {
	c := &conn{
		netCon: netConn,
		w:      netConn,
		r:      netConn,
		Closer: netConn,
		Config: conf,
	}
	if conf.CypherPool != nil {
		if conf.CypherPool.Reader != nil {
			c.r = conf.CypherPool.Reader(netConn)
		}
		if conf.CypherPool.Writer != nil {
			c.w = conf.CypherPool.Writer(netConn)
		}
	}
	return c
}

func (c *conn) resetDeadline() {
	err := c.netCon.SetDeadline(time.Now().Add(c.Config.IdleTimeout))
	if err != nil {
		log.Printf("socketman: SetDeadline failed: %s", err)
	}
}

func (c *conn) Write(b []byte) (n int, err error) {
	n, err = c.w.Write(b)
	if n > 0 && c.Config.IdleTimeout != 0 {
		c.resetDeadline()
	}
	return n, err
}

func (c *conn) Read(b []byte) (n int, err error) {
	n, err = c.r.Read(b)
	if n > 0 && c.Config.IdleTimeout != 0 {
		c.resetDeadline()
	}
	return n, err
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
