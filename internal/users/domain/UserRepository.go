package domain

import "context"

type UserRepository interface {
	Create(ctx context.Context, user *User) (*User, error)
	GetByID(ctx context.Context, id int32) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	List(ctx context.Context, limit, offset int) ([]*User, error)
	Count(ctx context.Context) (int, error)
	Update(ctx context.Context, user *User) error
	Deactivate(ctx context.Context, id int32) error
	ListGeneros(ctx context.Context) ([]*Genero, error)
}
