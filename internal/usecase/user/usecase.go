package user

import (
	"context"
	"errors"

	"user-management/internal/auth"
	userdomain "user-management/internal/domain/user"
	userrepo "user-management/internal/repository/user"
)

type Usecase struct {
	repo      userrepo.Repository
	jwtSecret string
}

func New(repo userrepo.Repository, secret string) *Usecase {
	return &Usecase{repo: repo, jwtSecret: secret}
}

func (uc *Usecase) Register(ctx context.Context, name, email, pw string) (*userdomain.User, error) {
	switch _, err := uc.repo.GetByEmail(ctx, email); {
	case err == nil:
		return nil, userdomain.ErrEmailExists
	case !errors.Is(err, userdomain.ErrNotFound):
		return nil, err
	}

	hashed, err := auth.HashPassword(pw)
	if err != nil {
		return nil, err
	}
	u := &userdomain.User{Name: name, Email: email, Password: hashed}
	if err := uc.repo.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (uc *Usecase) Login(ctx context.Context, email, pw string) (string, error) {
	u, err := uc.repo.GetByEmail(ctx, email)
	if err != nil {
		return "", userdomain.ErrInvalidCreds
	}
	if !auth.CheckPassword(u.Password, pw) {
		return "", userdomain.ErrInvalidCreds
	}
	return auth.GenerateToken(u.ID, uc.jwtSecret)
}

func (uc *Usecase) GetByID(ctx context.Context, id string) (*userdomain.User, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *Usecase) List(ctx context.Context) ([]*userdomain.User, error) {
	return uc.repo.List(ctx)
}

func (uc *Usecase) Update(ctx context.Context, id, name, email string) (*userdomain.User, error) {
	u, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if email != u.Email {
		switch other, err := uc.repo.GetByEmail(ctx, email); {
		case err == nil && other.ID != id:
			return nil, userdomain.ErrEmailExists
		case err != nil && !errors.Is(err, userdomain.ErrNotFound):
			return nil, err
		}
	}

	u.Name, u.Email = name, email
	if err := uc.repo.Update(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (uc *Usecase) Delete(ctx context.Context, id string) error {
	return uc.repo.Delete(ctx, id)
}
