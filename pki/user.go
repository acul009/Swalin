package pki

import "fmt"

func CreateAndApplyCurrentUserCert(username string, userPassword []byte, caPassword []byte) error {
	cert, key, err := GetCa(caPassword)
	if err != nil {
		return fmt.Errorf("failed to load CA: %v", err)
	}

	userCert, userKey, err := generateUserCert(username, key, cert)
	if err != nil {
		return fmt.Errorf("failed to generate user certificate: %v", err)
	}

	err = SaveCurrentCertAndKey(userCert, userKey, userPassword)
	if err != nil {
		return fmt.Errorf("failed to save current cert and key: %v", err)
	}

	return nil
}
