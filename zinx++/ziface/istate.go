package ziface

import "context"

type IStateManager interface {
	SetState(ctx context.Context, key string, value []byte, expiration int64) error

	GetState(ctx context.Context, key string) ([]byte, error)

	DeleteState(ctx context.Context, key string) error

	ExistsState(ctx context.Context, key string) (bool, error)
}
