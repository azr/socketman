package socketman

import (
	"crypto/cipher"
	"io"
)

//CypherPool will allow you to tell your client/server how to setup
//in house encryption.
//Encryption works on top of TLS for double noise.
type CypherPool interface {
	//Reader returns a new instatiation of a reader that knows how to decrypt from reader.
	//nil means none
	Reader(io.Reader) *cipher.StreamReader

	//Reader returns a new instatiation of a writer that knows how to encrypt to writer.
	//nil means none
	Writer(io.Writer) *cipher.StreamWriter

	//TODO: add possibility to recycle streams
}
