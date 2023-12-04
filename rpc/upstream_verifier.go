package rpc

import (
	"context"
	"fmt"

	"github.com/rahn-it/svalin/pki"
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
		chain, err := session.Verifier().VerifyPublicKey(c.Key)
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
		chain, err := session.Verifier().Verify(c.Cert)
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
	root     *pki.Certificate
	upstream *pki.Certificate
	ep       *RpcEndpoint
}

func NewUpstreamVerify(ep *RpcEndpoint) (*upstreamVerify, error) {

	return &upstreamVerify{
		ep: ep,
	}, nil
}

func (v *upstreamVerify) Verify(cert *pki.Certificate) ([]*pki.Certificate, error) {
	if cert == nil {
		return nil, fmt.Errorf("certificate is nil")
	}

	if v.root.Equal(cert) {
		return []*pki.Certificate{v.root}, nil
	}

	if v.upstream.Equal(cert) {
		// TODO: verify chain to root
		return []*pki.Certificate{v.upstream}, nil
	}

	chain := make([]*pki.Certificate, 0, 1)

	err := v.ep.SendSyncCommand(context.Background(),
		&verifyCertificateChainCmd{
			Cert:  cert,
			chain: chain,
		},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to request certificate chain: %w", err)
	}

	return cert.VerifyChain(nil, pki.CreatePool(chain), true)
}

func (v *upstreamVerify) VerifyPublicKey(pub *pki.PublicKey) ([]*pki.Certificate, error) {
	if v.root.PublicKey().Equal(pub) {
		return []*pki.Certificate{v.root}, nil
	}

	if v.upstream.PublicKey().Equal(pub) {
		return v.Verify(v.upstream)
	}

	cmd := &verifyCertificateChainCmd{
		Key: pub,
	}

	err := v.ep.SendSyncCommand(context.Background(), cmd)

	if err != nil {
		return nil, fmt.Errorf("failed to request certificate chain: %w", err)
	}

	chain := cmd.chain

	cert := chain[0]

	return cert.VerifyChain(nil, pki.CreatePool(chain), true)
}
