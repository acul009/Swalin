package rmm

import (
	"rahnit-rmm/pki"
	"rahnit-rmm/rpc"
)

type Server struct {
	*rpc.RpcServer[*Dependencies]
}

func NewServer(listenAddr string, rpcCommands *rpc.CommandCollection, credentials *pki.PermanentCredentials) (*Server, error) {
	deps := &Dependencies{}

	server, err := rpc.NewRpcServer(listenAddr, rpcCommands, credentials, deps)
	if err != nil {
		return nil, err
	}

	s := &Server{
		RpcServer: server,
	}

	return s, nil
}
