package pki

func NewNilVerifier() Verifier {
	return &nilVerifier{}
}

type nilVerifier struct{}

func (v *nilVerifier) Verify(cert *Certificate) ([]*Certificate, error) {
	return nil, nil
}

func (v *nilVerifier) VerifyPublicKey(pub *PublicKey) ([]*Certificate, error) {
	return nil, nil
}
