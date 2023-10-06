package pki_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
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
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	data := testData{
		Text: "test",
		Num:  1,
		Flt:  1.0,
	}

	marshalled, err := pki.MarshalAndSign(data, key, &key.PublicKey)
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
