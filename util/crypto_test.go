package util_test

import (
	"bytes"
	"github.com/rahn-it/svalin/util"
	"testing"
)

func TestEncryptDataWithPassword(t *testing.T) {
	password := []byte("password")

	data := []byte("Hello, I'm some data that I want to encrypt!")

	encryptedData, err := util.EncryptDataWithPassword(password, data)
	if err != nil {
		t.Fatalf("failed encrypting data: %v", err)
	}

	decryptedData, err := util.DecryptDataWithPassword(password, encryptedData)
	if err != nil {
		t.Fatalf("failed decrypting data: %v", err)
	}

	if !bytes.Equal(decryptedData, data) {
		t.Fatalf("decrypted data does not match original data")
	}

}
