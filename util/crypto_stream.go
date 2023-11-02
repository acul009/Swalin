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
	stream    io.ReadWriteCloser
	encryptor cipher.Stream
	decryptor cipher.Stream
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

	// Create a cipher cipher for encryption
	encryptor := cipher.NewCFBEncrypter(block, iv)
	decryptor := cipher.NewCFBDecrypter(block, iv)

	return &CryptoStream{
		stream:    stream,
		encryptor: encryptor,
		decryptor: decryptor,
	}, nil
}

func (c *CryptoStream) Read(dst []byte) (int, error) {
	src := make([]byte, len(dst))
	n, err := c.stream.Read(src)
	if n == 0 {
		return 0, err
	}

	c.encryptor.XORKeyStream(dst, src)
	return n, err
}

func (c *CryptoStream) Write(src []byte) (int, error) {
	dst := make([]byte, len(src))
	c.decryptor.XORKeyStream(dst, src)
	return c.stream.Write(dst)
}

func (c *CryptoStream) Close() error {
	return c.stream.Close()
}
