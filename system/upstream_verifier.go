package system

import (
	"context"
	"crypto/x509"
	"fmt"

	"github.com/rahn-it/svalin/pki"
	"github.com/rahn-it/svalin/rpc"
)

func CreateUpstreamVerificationCommandHandler(verifier pki.Verifier) rpc.RpcCommandHandler {
	return func() rpc.RpcCommand {
		return &requestKeyVerificationChainCommand{
			serverVerifier: verifier,
		}
	}
}

var _ pki.Verifier = (*upstreamVerifier)(nil)

type upstreamVerifier struct {
	ep              *rpc.RpcEndpoint
	upstream        *pki.Certificate
	root            *pki.Certificate
	rootPool        *x509.CertPool
	revocationStore *RevocationStore
}

func (v *upstreamVerifier) Verify(cert *pki.Certificate) ([]*pki.Certificate, error) {
	chain, err := v.VerifyPublicKey(cert.PublicKey())
	if err != nil {
		return nil, fmt.Errorf("failed to verify public key: %w", err)
	}

	if !chain[0].Equal(cert) {
		return nil, fmt.Errorf("received different certificate with same public key")
	}

	return chain, nil
}

func (v *upstreamVerifier) VerifyPublicKey(pub *pki.PublicKey) ([]*pki.Certificate, error) {
	if v.root.PublicKey().Equal(pub) {
		return []*pki.Certificate{v.root}, nil
	}

	if v.upstream.PublicKey().Equal(pub) {
		return []*pki.Certificate{v.upstream, v.root}, nil
	}

	cmd := &requestKeyVerificationChainCommand{
		Key: pub,
	}

	err := v.ep.SendSyncCommand(context.Background(), cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to request certificate chain: %w", err)
	}

	chain := cmd.getChain()

	if !chain[0].PublicKey().Equal(pub) {
		return nil, fmt.Errorf("server returned chain for wrong key")
	}

	intermediates := x509.NewCertPool()
	for _, cert := range chain[1 : len(chain)-1] {
		intermediates.AddCert(cert.ToX509())
	}

	verifiedChain, err := chain[0].VerifyChain(v.rootPool, intermediates)
	if err != nil {
		return nil, fmt.Errorf("failed to verify certificate chain: %w", err)
	}

	for _, cert := range verifiedChain {
		err := v.revocationStore.CheckCertificate(cert)
		if err != nil {
			return nil, fmt.Errorf("failed to ensure chain was not revoked: %w", err)
		}
	}

	return verifiedChain, nil
}

var _ rpc.RpcCommand = (*requestKeyVerificationChainCommand)(nil)

type requestKeyVerificationChainCommand struct {
	Key            *pki.PublicKey
	serverVerifier pki.Verifier
	chain          []*pki.Certificate
}

func (c *requestKeyVerificationChainCommand) ExecuteClient(session *rpc.RpcSession) error {
	chain := make([]*pki.Certificate, 0, 3)
	err := rpc.ReadMessage(session, &chain)
	if err != nil {
		return fmt.Errorf("error reading message: %w", err)
	}

	if len(chain) < 2 {
		return fmt.Errorf("certificate chain is too short")
	}

	c.chain = chain
	return nil
}

func (c *requestKeyVerificationChainCommand) ExecuteServer(session *rpc.RpcSession) error {
	chain, err := c.serverVerifier.VerifyPublicKey(c.Key)
	if err != nil {
		session.WriteResponseHeader(rpc.SessionResponseHeader{
			Code: 500,
			Msg:  "Unable to verify requested public key",
		})
		return fmt.Errorf("error verifying requested public key: %w", err)
	}

	session.WriteResponseHeader(rpc.SessionResponseHeader{
		Code: 200,
		Msg:  "OK",
	})

	err = rpc.WriteMessage[[]*pki.Certificate](session, chain)
	if err != nil {
		return fmt.Errorf("error writing message: %w", err)
	}

	return nil
}

func (c *requestKeyVerificationChainCommand) GetKey() string {
	return "verify-key"
}

func (c *requestKeyVerificationChainCommand) getChain() []*pki.Certificate {
	return c.chain
}
