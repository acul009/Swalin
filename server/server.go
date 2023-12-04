package server

import (
	"fmt"

	"github.com/rahn-it/svalin/config"
	"github.com/rahn-it/svalin/rmm"
	"github.com/rahn-it/svalin/rpc"
	"github.com/rahn-it/svalin/system"
	"github.com/rahn-it/svalin/util"

	"github.com/google/uuid"
)

type Server struct {
	*rpc.RpcServer
	serverConfig    *serverConfig
	profile         *config.Profile
	userStore       *userStore
	deviceStore     *deviceStore
	revocationStore *system.RevocationStore
	verifier        *LocalCertificateVerifier
	devices         util.ObservableMap[string, *system.DeviceInfo]
	configManager   *ConfigManager
}

func Open(profile *config.Profile) (*Server, error) {
	config := profile.Config()
	config.Default("server.address", "localhost:1234")

	scope := profile.Scope()

	serverConfig, err := openServerConfig(scope.Scope("server"))
	if err != nil {
		return nil, fmt.Errorf("error opening server config: %w", err)
	}

	userStore, err := openUserStore(scope.Scope("users"))
	if err != nil {
		return nil, fmt.Errorf("error opening user store: %w", err)
	}

	deviceStore, err := openDeviceStore(scope.Scope("devices"))
	if err != nil {
		return nil, fmt.Errorf("error opening device store: %w", err)
	}

	revocationStore, err := system.OpenRevocationStore(scope.Scope("revocation"), serverConfig.Root())
	if err != nil {
		return nil, fmt.Errorf("error opening revocation store: %w", err)
	}

	verifier, err := newLocalCertificateVerifier(serverConfig.Root(), userStore, deviceStore, revocationStore)
	if err != nil {
		return nil, fmt.Errorf("error creating local certificate verifier: %w", err)
	}

	ConfigManager := NewConfigManager(verifier, nil)

	devices := newDeviceList(deviceStore)

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
		RpcServer:       rpcS,
		profile:         profile,
		userStore:       userStore,
		deviceStore:     deviceStore,
		revocationStore: revocationStore,
		verifier:        verifier,
		devices:         devices,
		serverConfig:    serverConfig,
		configManager:   ConfigManager,
	}

	return s, nil
}

func (s *Server) Run() error {
	return s.RpcServer.Run()
}

func InitServer(profile *config.Profile) error {
	scope := profile.Scope()

	listenAddr := profile.Config().String("server.address")
	credentials, root, err := rpc.WaitForServerSetup(listenAddr)
	if err != nil {
		return fmt.Errorf("error waiting for server setup: %w", err)
	}

	err = initServerConfig(scope.Scope("server"), credentials, root)
	if err != nil {
		return fmt.Errorf("error initializing server config: %w", err)
	}

	return nil
}
