package server

import (
	"fmt"

	"github.com/rahn-it/svalin/config"
	"github.com/rahn-it/svalin/pki"
	"github.com/rahn-it/svalin/rmm"
	"github.com/rahn-it/svalin/rpc"
	"github.com/rahn-it/svalin/system"
	"github.com/rahn-it/svalin/util"

	"github.com/google/uuid"
)

type Server struct {
	*rpc.RpcServer
	serverConfig  *serverConfig
	profile       *config.Profile
	devices       util.ObservableMap[string, *system.DeviceInfo]
	configManager *ConfigManager
}

func Open(profile *config.Profile) (*Server, error) {
	config := profile.Config()
	config.Default("server.address", "localhost:1234")

	scope := profile.Scope()

	serverConfig, err := openServerConfig(scope.Scope("server"))
	if err != nil {
		return nil, fmt.Errorf("error opening server config: %w", err)
	}

	verifier, err := pki.NewLocalVerify()
	if err != nil {
		return nil, fmt.Errorf("error creating local verify: %w", err)
	}

	ConfigManager := NewConfigManager(verifier, nil)

	devices := NewDeviceList(scope.Scope("devices"))

	cmds := rpc.NewCommandCollection(
		rpc.PingHandler,
		rpc.RegisterUserHandler,
		// rpc.GetPendingEnrollmentsHandler,
		rpc.EnrollAgentHandler,
		rmm.CreateGetDevicesCommandHandler(devices),
		rpc.ForwardCommandHandler,
		rpc.VerifyCertificateChainHandler,
		// CreateHostConfigCommandHandler[*TunnelConfig],
	)

	listenAddr := config.String("server.address")

	rpcS, err := rpc.NewRpcServer(listenAddr, cmds, verifier, serverConfig.Credentials())
	if err != nil {
		return nil, fmt.Errorf("error creating rpc server: %w", err)
	}

	rpcS.Connections().Subscribe(
		func(_ uuid.UUID, rc *rpc.RpcConnection) {
			key := rc.Partner().PublicKey().Base64Encode()
			devices.setOnlineStatus(key, true)
		},
		func(_ uuid.UUID, rc *rpc.RpcConnection) {
			key := rc.Partner().PublicKey().Base64Encode()
			devices.setOnlineStatus(key, false)
		},
	)

	s := &Server{
		RpcServer:     rpcS,
		devices:       devices,
		serverConfig:  serverConfig,
		configManager: ConfigManager,
	}

	return s, nil
}

func (s *Server) Run() error {
	return s.RpcServer.Run()
}
