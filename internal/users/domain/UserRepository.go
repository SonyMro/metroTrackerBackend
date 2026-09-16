package domain

import "context"

type UserRepository interface {
	Crete(ctx context.Context, user *User) (*User, error)
	GetById(ctx context.Context, id int32) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	List(ctx context.Context) ([]*User, error)
	Update(ctx context.Context, user *User) error
	Deactivate(ctx context.Context, id int32) error
}
