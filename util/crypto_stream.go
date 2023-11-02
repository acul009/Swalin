package util

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"io"
)

// TODO: check if secure against padding attacks

type CryptoStream struct {
	stream io.ReadWriteCloser
	io.Reader
	io.Writer
}

func NewCryptoStream(stream io.ReadWriteCloser, key []byte, iv []byte) (*CryptoStream, error) {
	if len(key) < 32 {
		return nil, fmt.Errorf("key is too short")
	}

	if len(key) > 32 {
		// Create Hash if key is too long
		hasher := crypto.SHA256.New()
		n, err := hasher.Write(key)
		if err != nil {
			return nil, fmt.Errorf("failed to hash longer key: %w", err)
		}
		if n != len(key) {
			return nil, fmt.Errorf("failed to hash data: short write")
		}

		key = hasher.Sum(nil)
	}

	// Create a new AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed creating cipher: %w", err)
	}

	return &CryptoStream{
		stream: stream,
		Reader: cipher.StreamReader{
			S: cipher.NewCFBDecrypter(block, iv),
			R: stream,
		},
		Writer: cipher.StreamWriter{
			S: cipher.NewCFBEncrypter(block, iv),
			W: stream,
		},
	}, nil
}

func (c *CryptoStream) Close() error {
	return c.stream.Close()
}
