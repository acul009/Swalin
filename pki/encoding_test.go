package pki_test

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/rahn-it/svalin/pki"
)

type testMarshal struct {
	PublicKey   *pki.PublicKey
	Certificate *pki.Certificate
}

func TestJson(t *testing.T) {
	generatedKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	pub, err := pki.ImportPublicKey(generatedKey.PublicKey)
	if err != nil {
		t.Fatal(err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Your Organization"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0), // Valid for 1 year
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &generatedKey.PublicKey, generatedKey)
	if err != nil {
		t.Fatal(err)
	}

	cert, err := pki.CertificateFromBinary(derBytes)
	if err != nil {
		t.Fatal(err)
	}

	test := testMarshal{
		PublicKey:   pub,
		Certificate: cert,
	}

	marshalled, err := json.Marshal(test)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("%s\n", marshalled)

	unmarshalled := &testMarshal{}

	fmt.Printf("%+v\n", unmarshalled)

	err = json.Unmarshal(marshalled, unmarshalled)
	if err != nil {
		t.Fatal(err)
	}

	if !pub.Equal(unmarshalled.PublicKey) {
		t.Errorf("expected %v, got %v", pub, unmarshalled.PublicKey)
	}

	if !bytes.Equal(cert.BinaryEncode(), unmarshalled.Certificate.BinaryEncode()) {
		t.Errorf("expected %v, got %v", cert.BinaryEncode(), unmarshalled.Certificate.BinaryEncode())
	}

}
