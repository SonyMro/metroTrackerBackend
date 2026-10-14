package application

import (
	"context"
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"
	"metry.uk/provext4/MetroTrackerBackend/internal/users/domain"
)

type fakeUserRepository struct {
	getByEmail func(context.Context, string) (*domain.User, error)
	create     func(context.Context, *domain.User) (*domain.User, error)
}

// ListGeneros implements [domain.UserRepository].
func (f fakeUserRepository) ListGeneros(ctx context.Context) ([]*domain.Genero, error) {
	panic("unimplemented")
}

func (f fakeUserRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	return f.create(ctx, user)
}
func (f fakeUserRepository) GetByID(context.Context, int32) (*domain.User, error) {
	return nil, domain.ErrUserNotFound
}
func (f fakeUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return f.getByEmail(ctx, email)
}
func (f fakeUserRepository) Update(context.Context, *domain.User) error { return nil }
func (f fakeUserRepository) Deactivate(context.Context, int32) error    { return nil }

func TestCreateUserExecute(t *testing.T) {
	infraErr := errors.New("database unavailable")
	tests := []struct {
		name        string
		user        *domain.User
		emailResult *domain.User
		emailErr    error
		createErr   error
		want        error
		verify      func(*testing.T, *domain.User)
	}{
		{name: "normaliza y hashea", user: &domain.User{Nombre: " Ana ", Apepat: " Pérez ", Apemat: " Díaz ", Correo: " ANA@EXAMPLE.COM ", Password: "password8", IDGenero: 1}, emailErr: domain.ErrUserNotFound, verify: func(t *testing.T, user *domain.User) {
			if user.Correo != "ana@example.com" || user.Nombre != "Ana" || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("password8")) != nil {
				t.Fatal("usuario no normalizado o contraseña sin hash")
			}
		}},
		{name: "contraseña corta", user: &domain.User{Nombre: "Ana", Apepat: "Pérez", Correo: "a@b.c", Password: "short", IDGenero: 1}, want: domain.ErrInvalidPassword},
		{name: "correo duplicado", user: &domain.User{Nombre: "Ana", Apepat: "Pérez", Correo: "a@b.c", Password: "password8", IDGenero: 1}, emailResult: &domain.User{}, want: domain.ErrEmailAlreadyExists},
		{name: "propaga infraestructura", user: &domain.User{Nombre: "Ana", Apepat: "Pérez", Correo: "a@b.c", Password: "password8", IDGenero: 1}, emailErr: infraErr, want: infraErr},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			created := false
			repo := fakeUserRepository{getByEmail: func(context.Context, string) (*domain.User, error) { return tt.emailResult, tt.emailErr }, create: func(_ context.Context, user *domain.User) (*domain.User, error) {
				created = true
				if tt.verify != nil {
					tt.verify(t, user)
				}
				return user, tt.createErr
			}}
			_, err := NewCreateUser(repo).Execute(context.Background(), tt.user)
			if !errors.Is(err, tt.want) {
				t.Fatalf("error = %v, quiere %v", err, tt.want)
			}
			if tt.want != nil && created {
				t.Fatal("no debía persistir el usuario")
			}
		})
	}
}
