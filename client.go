package socketman

import (
	"crypto/tls"
	"io"
	"net"
)

//Client is a socket client
type Client struct {
	//Config is a configuration for new incoming connections
	Config
}

//Connect opens a tcp connection on server behind addr and calls handler.
//
//connection will be closed after the handler returns
//
//The syntax of addr is "host:port", like "127.0.0.1:8080".
//If host is omitted, as in ":8080".
//See net.Dial and tls.Dial for more details about address syntax.
func (c *Client) Connect(addr string, handler Handler) error {

	var con net.Conn
	var err error
	if c.Config.TLSConfig != nil {
		config := cloneTLSClientConfig(c.Config.TLSConfig)
		con, err = tls.Dial("tcp", addr, config)
	} else {
		con, err = net.Dial("tcp", addr)
	}
	if err != nil {
		return err
	}
	conn := &conn{
		netCon: con,
		Config: c.Config,
	}
	handler.ServeSocket(conn)
	return conn.netCon.Close()
}

//ConnectFunc calls Connect
func (c *Client) ConnectFunc(addr string, handler func(io.ReadWriter)) error {
	return c.Connect(addr, HandlerFunc(handler))
}
