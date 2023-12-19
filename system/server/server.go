package server

import (
	"fmt"
	"log"

	"github.com/rahn-it/svalin/config"
	"github.com/rahn-it/svalin/rpc"
	"github.com/rahn-it/svalin/system"
	"github.com/rahn-it/svalin/util"
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
	log.Printf("opening server for profile %s", profile.Name())

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

	userVerifier, error := newNewUserVerifier(serverConfig.Root())
	if error != nil {
		return nil, fmt.Errorf("error creating new user verifier: %w", error)
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

	// ConfigManager := NewConfigManager(verifier, nil)

	// devices := newDeviceList(deviceStore)

	cmds := rpc.NewCommandCollection(
		rpc.PingHandler,
		// rmm.CreateGetDevicesCommandHandler(devices),
		rpc.ForwardCommandHandler,
		system.CreateUpstreamVerificationCommandHandler(verifier),
		system.CreateRegisterUserCommandHandler(userVerifier, userStore.newUser),
	)

	listenAddr := config.String("server.address")

	rpcS, err := rpc.NewRpcServer(listenAddr, cmds, verifier, serverConfig.Credentials(), serverConfig.Root())
	if err != nil {
		return nil, fmt.Errorf("error creating rpc server: %w", err)
	}

	cmds.Add(system.CreateGetEnrollmentsCommandHandler(rpcS.Enrollments()))

	cmds.Add(system.CreateEnrollDeviceCommandHandler(rpcS.Enrollments()))

	// rpcS.Connections().Subscribe(
	// 	func(_ uuid.UUID, rc *rpc.RpcConnection) {
	// 		key := rc.Partner().PublicKey().Base64Encode()
	// 		devices.setOnlineStatus(key, true)
	// 	},
	// 	func(_ uuid.UUID, rc *rpc.RpcConnection) {
	// 		key := rc.Partner().PublicKey().Base64Encode()
	// 		devices.setOnlineStatus(key, false)
	// 	},
	// )

	s := &Server{
		RpcServer:       rpcS,
		profile:         profile,
		userStore:       userStore,
		deviceStore:     deviceStore,
		revocationStore: revocationStore,
		verifier:        verifier,
		// devices:         devices,
		serverConfig: serverConfig,
		// configManager:   ConfigManager,
	}

	loginHandler := system.NewLoginHandler(userStore.getUserByName, serverConfig.Seed(), serverConfig.Root(), serverConfig.Credentials().Certificate())
	rpcS.LoginHandler(loginHandler.HandleLoginRequest)

	return s, nil
}

func (s *Server) Run() error {
	return s.RpcServer.Run()
}

func Init(profile *config.Profile) error {
	scope := profile.Scope().Scope("server")

	configFound, err := checkForServerConfig(scope)
	if err != nil {
		return fmt.Errorf("error checking for server config: %w", err)
	}

	if configFound {
		return nil
	}

	profile.Config().Default("server.address", "localhost:1234")
	listenAddr := profile.Config().String("server.address")

	log.Printf("Server Waiting for initial setup on %s", listenAddr)

	credentials, root, err := rpc.WaitForServerSetup(listenAddr)
	if err != nil {
		return fmt.Errorf("error waiting for server setup: %w", err)
	}
	profile.Config().Save("server.address", listenAddr)

	err = initServerConfig(scope, credentials, root)
	if err != nil {
		return fmt.Errorf("error initializing server config: %w", err)
	}

	log.Printf("server setup complete for profile %s", profile.Name())

	return nil
}
