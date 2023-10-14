package pki

import "fmt"

func CreateAndApplyCurrentUserCert(username string, userPassword []byte, caPassword []byte) error {
	cert, key, err := GetRoot(caPassword)
	if err != nil {
		return fmt.Errorf("failed to load CA: %w", err)
	}

	userCert, userKey, err := generateUserCert(username, key, cert)
	if err != nil {
		return fmt.Errorf("failed to generate user certificate: %w", err)
	}

	err = SaveCurrentCertAndKey(userCert, userKey, userPassword)
	if err != nil {
		return fmt.Errorf("failed to save current cert and key: %w", err)
	}

	return nil
}
