package user

import (
	"context"
	"errors"
	"testing"

	"user-management/internal/auth"
	userdomain "user-management/internal/domain/user"
)

// fakeRepo is an in-memory userrepo.Repository for usecase tests.
type fakeRepo struct {
	users  map[string]*userdomain.User
	nextID int
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{users: map[string]*userdomain.User{}}
}

func (f *fakeRepo) Create(_ context.Context, u *userdomain.User) error {
	f.nextID++
	u.ID = string(rune('a' + f.nextID))
	f.users[u.ID] = u
	return nil
}

func (f *fakeRepo) GetByID(_ context.Context, id string) (*userdomain.User, error) {
	u, ok := f.users[id]
	if !ok {
		return nil, userdomain.ErrNotFound
	}
	return u, nil
}

func (f *fakeRepo) GetByEmail(_ context.Context, email string) (*userdomain.User, error) {
	for _, u := range f.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, userdomain.ErrNotFound
}

func (f *fakeRepo) List(_ context.Context) ([]*userdomain.User, error) {
	out := []*userdomain.User{}
	for _, u := range f.users {
		out = append(out, u)
	}
	return out, nil
}

func (f *fakeRepo) Update(_ context.Context, u *userdomain.User) error {
	if _, ok := f.users[u.ID]; !ok {
		return userdomain.ErrNotFound
	}
	f.users[u.ID] = u
	return nil
}

func (f *fakeRepo) Delete(_ context.Context, id string) error {
	if _, ok := f.users[id]; !ok {
		return userdomain.ErrNotFound
	}
	delete(f.users, id)
	return nil
}

func (f *fakeRepo) Count(_ context.Context) (int64, error) {
	return int64(len(f.users)), nil
}

func TestRegisterRejectsDuplicateEmail(t *testing.T) {
	uc := New(newFakeRepo(), "secret")
	ctx := context.Background()

	if _, err := uc.Register(ctx, "Ann", "ann@example.com", "password123"); err != nil {
		t.Fatalf("first register: %v", err)
	}
	_, err := uc.Register(ctx, "Ann Again", "ann@example.com", "password123")
	if !errors.Is(err, userdomain.ErrEmailExists) {
		t.Fatalf("want ErrEmailExists, got %v", err)
	}
}

func TestRegisterStoresHashedPassword(t *testing.T) {
	repo := newFakeRepo()
	uc := New(repo, "secret")

	u, err := uc.Register(context.Background(), "Ann", "ann@example.com", "password123")
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if u.Password == "password123" {
		t.Fatal("password stored in plain text")
	}
	if !auth.CheckPassword(u.Password, "password123") {
		t.Fatal("stored hash does not match the original password")
	}
}

func TestLoginWrongPassword(t *testing.T) {
	uc := New(newFakeRepo(), "secret")
	ctx := context.Background()

	if _, err := uc.Register(ctx, "Ann", "ann@example.com", "password123"); err != nil {
		t.Fatalf("register: %v", err)
	}
	if _, err := uc.Login(ctx, "ann@example.com", "wrong-password"); !errors.Is(err, userdomain.ErrInvalidCreds) {
		t.Fatalf("want ErrInvalidCreds, got %v", err)
	}
}

func TestLoginReturnsUsableToken(t *testing.T) {
	uc := New(newFakeRepo(), "secret")
	ctx := context.Background()

	u, err := uc.Register(ctx, "Ann", "ann@example.com", "password123")
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	token, err := uc.Login(ctx, "ann@example.com", "password123")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	userID, err := auth.ParseToken(token, "secret")
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}
	if userID != u.ID {
		t.Fatalf("token subject = %q, want %q", userID, u.ID)
	}
}

func TestUpdateRejectsEmailTakenByAnotherUser(t *testing.T) {
	uc := New(newFakeRepo(), "secret")
	ctx := context.Background()

	first, err := uc.Register(ctx, "Ann", "ann@example.com", "password123")
	if err != nil {
		t.Fatalf("register ann: %v", err)
	}
	if _, err := uc.Register(ctx, "Bob", "bob@example.com", "password123"); err != nil {
		t.Fatalf("register bob: %v", err)
	}
	if _, err := uc.Update(ctx, first.ID, "Ann", "bob@example.com"); !errors.Is(err, userdomain.ErrEmailExists) {
		t.Fatalf("want ErrEmailExists, got %v", err)
	}
}
