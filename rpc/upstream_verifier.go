package rpc

import (
	"context"
	"crypto/x509"
	"fmt"
	"rahnit-rmm/pki"
)

func VerifyCertificateChainHandler() RpcCommand {
	return &verifyCertificateChainCmd{}
}

type verifyCertificateChainCmd struct {
	Key   *pki.PublicKey
	Cert  *pki.Certificate
	chain []*pki.Certificate
}

func (c *verifyCertificateChainCmd) ExecuteServer(session *RpcSession) error {
	if c.Key != nil {
		chain, err := session.connection.verifier.VerifyPublicKey(c.Key)
		if err != nil {
			session.WriteResponseHeader(SessionResponseHeader{
				Code: 500,
				Msg:  "Internal Server Error",
			})
			return fmt.Errorf("error verifying public key: %w", err)
		}

		session.WriteResponseHeader(SessionResponseHeader{
			Code: 200,
			Msg:  "OK",
		})

		err = WriteMessage[[]*pki.Certificate](session, chain)
		if err != nil {
			return fmt.Errorf("error writing message: %w", err)
		}

		return nil
	}

	if c.Cert != nil {
		chain, err := session.connection.verifier.Verify(c.Cert)
		if err != nil {
			session.WriteResponseHeader(SessionResponseHeader{
				Code: 500,
				Msg:  "Internal Server Error",
			})
			return fmt.Errorf("error verifying certificate: %w", err)
		}

		session.WriteResponseHeader(SessionResponseHeader{
			Code: 200,
			Msg:  "OK",
		})

		err = WriteMessage[[]*pki.Certificate](session, chain)
		if err != nil {
			return fmt.Errorf("error writing message: %w", err)
		}

		return nil
	}

	session.WriteResponseHeader(SessionResponseHeader{
		Code: 400,
		Msg:  "Bad Request",
	})
	return fmt.Errorf("no certificate or public key specified")

}

func (c *verifyCertificateChainCmd) ExecuteClient(session *RpcSession) error {
	chain := make([]*pki.Certificate, 0)
	err := ReadMessage[[]*pki.Certificate](session, chain)
	if err != nil {
		return fmt.Errorf("error reading message: %w", err)
	}

	c.chain = chain
	return nil
}

func (c *verifyCertificateChainCmd) GetKey() string {
	return "verify-certificate-chain"
}

type upstreamVerify struct {
	rootPool *x509.CertPool
	ep       *RpcEndpoint
}

func NewUpstreamVerify(ep *RpcEndpoint) (*upstreamVerify, error) {
	rootCert, err := pki.Root.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to load root certificate: %w", err)
	}

	rootPool := x509.NewCertPool()
	rootPool.AddCert(rootCert.ToX509())

	return &upstreamVerify{
		rootPool: rootPool,
		ep:       ep,
	}, nil
}

func (v *upstreamVerify) options(intermediates []*pki.Certificate) x509.VerifyOptions {

	pool := x509.NewCertPool()
	for _, cert := range intermediates {
		pool.AddCert(cert.ToX509())
	}

	return x509.VerifyOptions{
		Roots:         v.rootPool,
		Intermediates: pool,
	}
}

func (v *upstreamVerify) Verify(cert *pki.Certificate) ([]*pki.Certificate, error) {
	if cert == nil {
		return nil, fmt.Errorf("certificate is nil")
	}

	root, err := pki.Root.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to check if certificate is root: %w", err)
	}
	if root.Equal(cert) {
		return []*pki.Certificate{root}, nil
	}

	upstream, err := pki.Upstream.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to check if certificate is upstream: %w", err)
	}
	if upstream.Equal(cert) {
		return v.Verify(upstream)
	}

	chain := make([]*pki.Certificate, 0, 1)

	err = v.ep.SendSyncCommand(context.Background(),
		&verifyCertificateChainCmd{
			Cert:  cert,
			chain: chain,
		},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to request certificate chain: %w", err)
	}

	chains, err := cert.ToX509().Verify(v.options(chain))
	if err != nil || len(chains) == 0 {
		return nil, fmt.Errorf("failed to verify certificate: %w", err)
	}

	for _, cert := range chains[0] {
		workingCert, err := pki.ImportCertificate(cert)
		if err != nil {
			return nil, fmt.Errorf("failed to import certificate: %w", err)
		}

		err = v.checkCertificateInfo(workingCert)
		if err != nil {
			return nil, fmt.Errorf("failed to check certificate info: %w", err)
		}

		chain = append(chain, workingCert)
	}

	return chain, nil
}

func (v *upstreamVerify) checkCertificateInfo(cert *pki.Certificate) error {
	err := v.checkRevoked(cert)
	if err != nil {
		return fmt.Errorf("certificate has been revoked: %w", err)
	}

	return nil
}

func (v *upstreamVerify) checkRevoked(cert *pki.Certificate) error {
	// TODO
	return nil
}

func (v *upstreamVerify) VerifyPublicKey(pub *pki.PublicKey) ([]*pki.Certificate, error) {
	root, err := pki.Root.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to check if public key is root: %w", err)
	}

	if root.GetPublicKey().Equal(pub) {
		return []*pki.Certificate{root}, nil
	}

	upstream, err := pki.Upstream.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to check if public key is upstream: %w", err)
	}
	if upstream.GetPublicKey().Equal(pub) {
		return v.Verify(upstream)
	}

	cmd := &verifyCertificateChainCmd{
		Key: pub,
	}

	err = v.ep.SendSyncCommand(context.Background(), cmd)

	if err != nil {
		return nil, fmt.Errorf("failed to request certificate chain: %w", err)
	}

	chain := cmd.chain

	cert := chain[0]

	verifiedChain := make([]*pki.Certificate, 0, len(chain))

	chains, err := cert.ToX509().Verify(v.options(chain))
	if err != nil || len(chains) == 0 {
		return nil, fmt.Errorf("failed to verify certificate: %w", err)
	}

	for _, cert := range chains[0] {
		workingCert, err := pki.ImportCertificate(cert)
		if err != nil {
			return nil, fmt.Errorf("failed to import certificate: %w", err)
		}

		err = v.checkCertificateInfo(workingCert)
		if err != nil {
			return nil, fmt.Errorf("failed to check certificate info: %w", err)
		}

		verifiedChain = append(verifiedChain, workingCert)
	}

	return verifiedChain, nil
}
