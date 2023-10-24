package pki_test

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"io"
	"rahnit-rmm/pki"
	"reflect"
	"testing"
)

type testData struct {
	Text string
	Num  int
	Flt  float64
}

func TestSignBytes(t *testing.T) {
	generatedKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	data := testData{
		Text: "test",
		Num:  1,
		Flt:  1.0,
	}

	key := pki.PrivateKey(*generatedKey)
	pubKeyGen := pki.PublicKey(key.PublicKey)

	marshalled, err := pki.MarshalAndSign(data, &key, &pubKeyGen)
	if err != nil {
		t.Fatal(err)
	}

	unmarshalled := &testData{}

	pub, err := pki.UnmarshalAndVerify(marshalled, unmarshalled)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(data, *unmarshalled) {
		t.Errorf("expected %v, got %v", data, unmarshalled)
	}

	if !reflect.DeepEqual(key.PublicKey, *pub) {
		t.Errorf("expected %v, got %v", key.PublicKey, pub)
	}
}

func TestPackedReadWrite(t *testing.T) {
	generatedKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	key := pki.PrivateKey(*generatedKey)
	pubKeyGen := pki.PublicKey(key.PublicKey)

	data1 := testData{
		Text: "test",
		Num:  1,
		Flt:  1.0,
	}

	data2 := testData{
		Text: "test2",
		Num:  2,
		Flt:  2.0,
	}

	data3 := make([]byte, 10000)
	for i := 0; i < len(data3); i++ {
		data3[i] = byte(i)
	}

	marshalled1, err := pki.MarshalAndSign(data1, &key, &pubKeyGen)
	if err != nil {
		t.Fatal(err)
	}

	marshalled2, err := pki.MarshalAndSign(data2, &key, &pubKeyGen)
	if err != nil {
		t.Fatal(err)
	}

	marshalled3, err := pki.MarshalAndSign(data3, &key, &pubKeyGen)
	if err != nil {
		t.Fatal(err)
	}

	marshalled := bytes.Join([][]byte{marshalled1, marshalled2, marshalled3}, nil)

	reader, writer := io.Pipe()

	errChan := make(chan error)

	go func() {
		defer writer.Close()
		_, err := writer.Write(marshalled)
		if err != nil {
			errChan <- err
		}
		close(errChan)
	}()

	unmarshalled1 := &testData{}

	unmarshalled2 := &testData{}

	unmarshalled3 := make([]byte, 10000)

	pub1, err := pki.ReadAndUnmarshalAndVerify(reader, unmarshalled1)
	if err != nil {
		t.Fatal(err)
	}

	pub2, err := pki.ReadAndUnmarshalAndVerify(reader, unmarshalled2)
	if err != nil {
		t.Fatal(err)
	}

	pub3, err := pki.ReadAndUnmarshalAndVerify(reader, &unmarshalled3)
	if err != nil {
		t.Fatal(err)
	}

	err, ok := <-errChan
	if ok {
		t.Fatal(err)
	}

	if !key.PublicKey.Equal(pub1) {
		t.Errorf("expected %v, got %v", key.PublicKey, pub1)
	}

	if !key.PublicKey.Equal(pub2) {
		t.Errorf("expected %v, got %v", key.PublicKey, pub2)
	}

	if !key.PublicKey.Equal(pub3) {
		t.Errorf("expected %v, got %v", key.PublicKey, pub3)
	}

	if !reflect.DeepEqual(data1, *unmarshalled1) {
		t.Errorf("expected %v, got %v", data1, unmarshalled1)
	}

	if !reflect.DeepEqual(data2, *unmarshalled2) {
		t.Errorf("expected %v, got %v", data2, unmarshalled2)
	}

	if !reflect.DeepEqual(data3, unmarshalled3) {
		t.Errorf("expected %v, got %v", data3, unmarshalled3)
	}
}
