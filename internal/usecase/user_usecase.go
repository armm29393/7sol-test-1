package usecase

import (
	"context"

	"user-management/internal/auth"
	"user-management/internal/domain"
)

type UserUsecase struct {
	repo      domain.UserRepository
	jwtSecret string
}

func NewUserUsecase(repo domain.UserRepository, secret string) *UserUsecase {
	return &UserUsecase{repo: repo, jwtSecret: secret}
}

func (uc *UserUsecase) Register(ctx context.Context, name, email, pw string) (*domain.User, error) {
	if _, err := uc.repo.GetByEmail(ctx, email); err == nil {
		return nil, domain.ErrEmailExists
	}
	hashed, err := auth.HashPassword(pw)
	if err != nil {
		return nil, err
	}
	u := &domain.User{Name: name, Email: email, Password: hashed}
	if err := uc.repo.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (uc *UserUsecase) Login(ctx context.Context, email, pw string) (string, error) {
	u, err := uc.repo.GetByEmail(ctx, email)
	if err != nil {
		return "", domain.ErrInvalidCreds
	}
	if !auth.CheckPassword(u.Password, pw) {
		return "", domain.ErrInvalidCreds
	}
	return auth.GenerateToken(u.ID, uc.jwtSecret)
}

// ... GetByID, List, Update, Delete ก็ห่อ repo เฉยๆ ตาม pattern เดียวกัน
