package pki

import (
	"fmt"
)

var Root = &storedCertificate{
	filename:      rootCertFilePath,
	allowOverride: false,
}

// Custom go error to indicate that the CA certificate is missing
type noRootCertError struct {
	cause error
}

func (e noRootCertError) Error() string {
	return fmt.Errorf("root certificate not found: %w", e.cause).Error()
}

func (e noRootCertError) Unwrap() error {
	return e.cause
}

var ErrNoRootCert = noRootCertError{}

func (e noRootCertError) Is(target error) bool {
	_, ok := target.(noRootCertError)
	return ok
}

const (
	rootCertFilePath = "root.crt"
)

func InitRoot(rootName string, password []byte) error {
	if Root.Available() {
		return fmt.Errorf("root certificate already exists")
	}

	rootCert, rootKey, err := generateRootCert(rootName)
	if err != nil {
		return fmt.Errorf("failed to generate root certificate: %w", err)
	}

	fmt.Printf("root certificate: %+v\n", rootCert)

	err = Root.Set(rootCert)
	if err != nil {
		return fmt.Errorf("failed to save root certificate: %w", err)
	}

	err = SaveUserCredentials(rootName, password, rootCert, rootKey)
	if err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}

	return nil
}
