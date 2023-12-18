package agent

import (
	"errors"
	"fmt"

	"github.com/rahn-it/svalin/db"
	"github.com/rahn-it/svalin/pki"
	"github.com/rahn-it/svalin/rpc"
	"github.com/rahn-it/svalin/system"
)

type agentConfig struct {
	scope       db.Scope
	root        *pki.Certificate
	upstream    *pki.Certificate
	credentials *pki.PermanentCredentials
	serverAddr  string
}

func openClientConfig(scope db.Scope) (*agentConfig, error) {
	conf := &agentConfig{
		scope: scope,
	}

	err := conf.loadRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to load root: %w", err)
	}

	err = conf.loadUpstream()
	if err != nil {
		return nil, fmt.Errorf("failed to load upstream: %w", err)
	}

	err = conf.loadCredentials()
	if err != nil {
		return nil, fmt.Errorf("failed to load credentials: %w", err)
	}

	err = conf.loadServerAddr()
	if err != nil {
		return nil, fmt.Errorf("failed to load server address: %w", err)
	}

	return conf, nil
}

func (conf *agentConfig) loadRoot() error {
	return conf.scope.View(func(b db.Bucket) error {
		raw := b.Get([]byte("root"))
		if raw == nil {
			return errors.New("root certificate not found")
		}
		root, err := pki.CertificateFromPem(raw)
		if err != nil {
			return fmt.Errorf("failed to load root certificate: %w", err)
		}

		conf.root = root
		return nil
	})
}

func (conf *agentConfig) loadUpstream() error {
	return conf.scope.View(func(b db.Bucket) error {
		raw := b.Get([]byte("upstream"))
		if raw == nil {
			return errors.New("upstream certificate not found")
		}
		upstream, err := pki.CertificateFromPem(raw)
		if err != nil {
			return fmt.Errorf("failed to load upstream certificate: %w", err)
		}

		conf.upstream = upstream
		return nil
	})
}

func (conf *agentConfig) loadCredentials() error {

	creds, err := system.LoadHostCredentials(conf.scope)
	if err != nil {
		return fmt.Errorf("failed to load host credentials: %w", err)
	}

	conf.credentials = creds
	return nil
}

func (conf *agentConfig) loadServerAddr() error {
	return conf.scope.View(func(b db.Bucket) error {
		addr := b.Get([]byte("server-addr"))
		if addr == nil {
			return errors.New("server address not found")
		}
		conf.serverAddr = string(addr)
		return nil
	})
}

func (conf *agentConfig) Root() *pki.Certificate {
	return conf.root
}

func (conf *agentConfig) Upstream() *pki.Certificate {
	return conf.upstream
}

func (conf *agentConfig) Credentials() *pki.PermanentCredentials {
	return conf.credentials
}

func (conf *agentConfig) ServerAddr() string {
	return conf.serverAddr
}

func checkForAgentConfig(scope db.Scope) (bool, error) {
	found := false
	err := scope.View(func(b db.Bucket) error {
		root := b.Get([]byte("root"))
		if root != nil {
			found = true
		}
		return nil
	})
	if err != nil {
		return false, fmt.Errorf("failed to check for agent config: %w", err)
	}

	return found, nil
}

func initAgentConfig(scope db.Scope, address string, initInfo *rpc.EndPointInitInfo) error {
	return scope.Update(func(b db.Bucket) error {
		err := b.Put([]byte("root"), initInfo.Root.PemEncode())
		if err != nil {
			return fmt.Errorf("failed to save root certificate: %w", err)
		}

		err = b.Put([]byte("upstream"), initInfo.Upstream.PemEncode())
		if err != nil {
			return fmt.Errorf("failed to save upstream certificate: %w", err)
		}

		err = system.SaveHostCredentials(b, initInfo.Credentials)
		if err != nil {
			return fmt.Errorf("failed to save host credentials: %w", err)
		}

		err = b.Put([]byte("serverAddr"), []byte(address))
		if err != nil {
			return fmt.Errorf("failed to save server address: %w", err)
		}

		return nil
	})
}
