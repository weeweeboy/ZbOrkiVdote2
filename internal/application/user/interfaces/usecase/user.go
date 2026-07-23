package usecase

import "context"

type UserUsecase interface {
	GetSborka(ctx context.Context, name string, position string) (string, error)
}
