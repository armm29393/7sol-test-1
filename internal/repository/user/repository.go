package user

import (
	"context"

	userdomain "user-management/internal/domain/user"
)

// Repository is the persistence contract for users. Implementations live
// alongside it in this package (see mongo.go).
type Repository interface {
	Create(ctx context.Context, user *userdomain.User) error
	GetByID(ctx context.Context, id string) (*userdomain.User, error)
	GetByEmail(ctx context.Context, email string) (*userdomain.User, error)
	List(ctx context.Context) ([]*userdomain.User, error)
	Update(ctx context.Context, user *userdomain.User) error
	Delete(ctx context.Context, id string) error
	Count(ctx context.Context) (int64, error)
}
