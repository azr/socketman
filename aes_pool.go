package socketman

import (
	"crypto/aes"
	"crypto/cipher"
	"io"
)

//NewAESPool instantiates a pool of aes encryptor/decryptor
func NewAESPool(key []byte) (*AESPool, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return &AESPool{
		block: block,
	}, nil
}

//AESPool will create aes stream writers and readers for you.
//TODO: recycle streams
type AESPool struct {
	block cipher.Block
}

//Reader will return a new StreamReader that can decode
//using AESPool key.
func (p *AESPool) Reader(r io.Reader) *cipher.StreamReader {
	// If the key is unique for each ciphertext, then it's ok to use a zero
	// IV.
	var iv [aes.BlockSize]byte
	stream := cipher.NewOFB(p.block, iv[:])
	return &cipher.StreamReader{S: stream, R: r}
}

//Writer will return a new StreamWriter that can decode
//using AESPool key.
func (p *AESPool) Writer(w io.Writer) *cipher.StreamWriter {
	// If the key is unique for each ciphertext, then it's ok to use a zero
	// IV.
	var iv [aes.BlockSize]byte
	stream := cipher.NewOFB(p.block, iv[:])
	return &cipher.StreamWriter{S: stream, W: w}
}

// func (p *AESPool) PutReader(r io.Reader) {
//   if r is a *cipher.StreamReader ...

// func (p *AESPool) PutWriter(r io.Writer) {
//   if r is a *cipher.StreamWriter ...
