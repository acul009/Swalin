package rmm

import (
	"fmt"

	"github.com/rahn-it/svalin/config"
	"github.com/rahn-it/svalin/pki"
	"github.com/rahn-it/svalin/rpc"
	"github.com/rahn-it/svalin/util"

	"github.com/google/uuid"
)

type Server struct {
	*rpc.RpcServer
	devices       util.UpdateableMap[string, *DeviceInfo]
	configManager *ConfigManager
}

func NewDefaultServer(listenAddr string, credentials *pki.PermanentCredentials) (*Server, error) {

	verifier, err := pki.NewLocalVerify()
	if err != nil {
		return nil, fmt.Errorf("error creating local verify: %w", err)
	}

	ConfigManager := NewConfigManager(verifier, config.DB())

	devices, err := NewDeviceListFromDB()
	if err != nil {
		return nil, fmt.Errorf("error loading devices from db: %w", err)
	}

	cmds := rpc.NewCommandCollection(
		rpc.PingHandler,
		rpc.RegisterUserHandler,
		// rpc.GetPendingEnrollmentsHandler,
		rpc.EnrollAgentHandler,
		CreateGetDevicesCommandHandler(devices),
		rpc.ForwardCommandHandler,
		rpc.VerifyCertificateChainHandler,
		// CreateHostConfigCommandHandler[*TunnelConfig],
	)

	rpcS, err := rpc.NewRpcServer(listenAddr, cmds, verifier, credentials)
	if err != nil {
		return nil, fmt.Errorf("error creating rpc server: %w", err)
	}

	rpcS.Connections().Subscribe(
		func(_ uuid.UUID, rc *rpc.RpcConnection) {
			key := rc.Partner().GetPublicKey().Base64Encode()
			devices.Update(key, func(d *DeviceInfo, found bool) (*DeviceInfo, bool) {
				if !found {
					return nil, false
				}

				d.Online = true
				return d, true
			})
		},
		func(_ uuid.UUID, rc *rpc.RpcConnection) {
			key := rc.Partner().GetPublicKey().Base64Encode()
			devices.Update(key, func(d *DeviceInfo, found bool) (*DeviceInfo, bool) {
				if !found {
					return nil, false
				}

				d.Online = false
				return d, true
			})
		},
	)

	s := &Server{
		RpcServer:     rpcS,
		devices:       devices,
		configManager: ConfigManager,
	}

	return s, nil
}

func (s *Server) Run() error {
	return s.RpcServer.Run()
}
