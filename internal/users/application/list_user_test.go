package application

import (
	"context"
	"testing"

	"metry.uk/provext4/MetroTrackerBackend/internal/users/domain"
)

func TestListUsersExecute(t *testing.T) {
	tests := []struct {
		name                                                         string
		page, size, total, wantPage, wantSize, wantOffset, wantPages int
	}{
		{"defaults", 0, 0, 41, 1, 20, 0, 3},
		{"limita máximo", 2, 101, 201, 2, 100, 100, 3},
		{"segunda página", 2, 20, 40, 2, 20, 20, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := fakeUserRepository{count: func(context.Context) (int, error) { return tt.total, nil }, list: func(_ context.Context, limit, offset int) ([]*domain.User, error) {
				if limit != tt.wantSize || offset != tt.wantOffset {
					t.Fatalf("limit/offset = %d/%d", limit, offset)
				}
				return []*domain.User{}, nil
			}}
			result, err := NewListUsers(repo).Execute(context.Background(), tt.page, tt.size)
			if err != nil || result.Page != tt.wantPage || result.PageSize != tt.wantSize || result.Total != tt.total || result.TotalPages != tt.wantPages {
				t.Fatalf("resultado inesperado: %#v, %v", result, err)
			}
		})
	}
}
