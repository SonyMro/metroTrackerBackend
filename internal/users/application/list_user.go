package application

import (
	"context"

	"metry.uk/provext4/MetroTrackerBackend/internal/users/domain"
)

type UserListResult struct {
	Items      []*domain.User
	Page       int
	PageSize   int
	Total      int
	TotalPages int
}

type ListUsers struct {
	repository domain.UserRepository
}

func NewListUsers(

	repository domain.UserRepository,

) *ListUsers {

	return &ListUsers{
		repository: repository,
	}

}

func (uc *ListUsers) Execute(

	ctx context.Context,
	page int,
	pageSize int,

) (*UserListResult, error) {

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	//offset := (page - 1) * pageSize
	items, err := uc.repository.List(
		ctx,
	)
	if err != nil {
		return nil, err
	}
	totalPages := 0

	return &UserListResult{
		Items:      items,
		Page:       page,
		PageSize:   pageSize,
		Total:      0,
		TotalPages: totalPages,
	}, nil

}
