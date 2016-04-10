package socketman

import (
	"crypto/tls"
	"time"
)

//Config is a socket configuration.
type Config struct {
	// TLS config for secure Sockets
	TLSConfig *tls.Config

	// After a connection is opened and after each successful I/O call
	// SetDeadline will be called upon that connection.
	//
	// A deadline is an absolute time after which I/O operations
	// fail with a timeout (see type net.Error) instead of
	// blocking. The deadline applies to all future I/O, not just
	// the immediately following call to Read or Write.
	//
	// A zero value means I/O operations will not time out.
	IdleDeadline time.Duration
}
