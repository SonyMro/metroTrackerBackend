package application

import (
	"context"
	"fmt"

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

func NewListUsers(repository domain.UserRepository) *ListUsers {
	return &ListUsers{repository: repository}
}

func (uc *ListUsers) Execute(ctx context.Context, page, pageSize int) (*UserListResult, error) {

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize
	total, err := uc.repository.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("contar usuarios: %w", err)
	}
	items, err := uc.repository.List(ctx, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("listar usuarios: %w", err)
	}
	totalPages := (total + pageSize - 1) / pageSize

	return &UserListResult{
		Items:      items,
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}
