package rpc

import (
	"context"
	"rahnit-rmm/pki"
	"rahnit-rmm/util"
)

type Dispatcher interface {
	SendCommand(ctx context.Context, cmd RpcCommand) (util.AsyncAction, error)
	SendSyncCommand(ctx context.Context, cmd RpcCommand) error
	SendCommandTo(ctx context.Context, to *pki.Certificate, cmd RpcCommand) (util.AsyncAction, error)
	SendSyncCommandTo(ctx context.Context, to *pki.Certificate, cmd RpcCommand) error
}
