package util_test

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"testing"
	"time"

	. "rahnit-rmm/util"

	"golang.org/x/crypto/chacha20poly1305"
)

type DoublePipe struct {
	io.Reader
	io.Writer
}

func new2WayPipe() (*DoublePipe, *DoublePipe) {
	// b1 := bytes.NewBuffer(make([]byte, 1024*1024))
	// b2 := bytes.NewBuffer(make([]byte, 1024*1024))
	// return &DoublePipe{b1, b2}, &DoublePipe{b2, b1}
	r1, w1 := io.Pipe()
	r2, w2 := io.Pipe()

	return &DoublePipe{r1, w2}, &DoublePipe{r2, w1}
}

func (p *DoublePipe) Close() error {
	return nil
}

func TestCryptoStream1M(t *testing.T) {
	key := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, key)
	if err != nil {
		t.Fatal(err)
	}

	testData := make([]byte, 1024*1024)
	_, err = io.ReadFull(rand.Reader, testData)
	if err != nil {
		t.Fatal(err)
	}

	pipe1, pipe2 := new2WayPipe()

	stream1, err := NewDefaultCipherStream(pipe1, key)
	if err != nil {
		t.Fatal(err)
	}

	stream2, err := NewDefaultCipherStream(pipe2, key)
	if err != nil {
		t.Fatal(err)
	}

	errChan := make(chan error)

	t.Logf("payload size: %d", len(testData))

	var written int
	go func() {
		var err error
		written, err = stream1.Write(testData)
		errChan <- err
	}()

	receive := make([]byte, len(testData))

	var read int
	go func() {
		var err error
		read, err = io.ReadFull(stream2, receive)
		errChan <- err
	}()

	go func() {
		time.Sleep(time.Second * 3)
		errChan <- fmt.Errorf("timeout")
	}()

	err = <-errChan
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("goroutine 1 finished")

	err = <-errChan
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("goroutine 2 finished")

	t.Logf("written: %d", written)
	t.Logf("read: %d", read)

	if !bytes.Equal(testData, receive) {
		t.Fatal("data mismatch")
	}
}

func TestCryptoStreamText(t *testing.T) {
	key := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, key)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("Test key ready")

	testData := []byte("This is a text message which should be encrypted and a bit longer to actually see this thing work")

	pipe1, pipe2 := new2WayPipe()

	stream1, err := NewDefaultCipherStream(pipe1, key)
	if err != nil {
		t.Fatal(err)
	}

	stream2, err := NewDefaultCipherStream(pipe2, key)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		_, err := stream1.Write(testData)
		if err != nil {
			t.Error(err)
		}
	}()

	receive := make([]byte, len(testData))
	_, err = io.ReadFull(stream2, receive)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("testData: ", string(testData))
	fmt.Println("receive: ", string(receive))

	if !bytes.Equal(testData, receive) {
		t.Fatal("data mismatch")
	}
}

func TestChaCha20(t *testing.T) {
	key := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, key)
	if err != nil {
		t.Fatal(err)
	}

	testData := []byte("This is a text message which should be encrypted and a bit longer to actually see this thing work")

	cipher, err := chacha20poly1305.NewX(key)
	if err != nil {
		t.Fatal(err)
	}

	nonce := make([]byte, cipher.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		t.Fatal(err)
	}

	encrypted := make([]byte, len(testData)+cipher.Overhead())
	encrypted = cipher.Seal(encrypted[:0], nonce, testData, nil)

	decrypted, err := cipher.Open(encrypted[:0], nonce, encrypted, nil)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(testData, decrypted) {
		t.Fatal("data mismatch")
	}
}
