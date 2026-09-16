package application

import (
	"context"

	"metry.uk/provext4/MetroTrackerBackend/internal/users/domain"
)

type GetUser struct {
	repository domain.UserRepository
}

func NewGetUser(
	repository domain.UserRepository,
) *GetUser {
	return &GetUser{
		repository: repository,
	}
}

func (uc *GetUser) Execute(
	ctx context.Context,
	id int32,
) (*domain.User, error) {
	if id <= 0 {
		return nil, domain.ErrUserNotFound
	}

	return uc.repository.GetById(
		ctx,
		id,
	)
}
