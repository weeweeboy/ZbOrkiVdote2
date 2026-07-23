package storage

import "context"

type UserStorage interface {
	GetH(ctx context.Context, name string) (int32, error)
	GetI(ctx context.Context, id int) (string, error)
}
