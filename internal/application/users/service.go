package users

import (
	"context"
	"fmt"

	domainusers "github.com/fastygo/app-gocms/internal/domain/users"
)

type Repository interface {
	SaveUser(context.Context, domainusers.User) error
	GetUser(context.Context, string) (domainusers.User, bool, error)
	ListUsers(context.Context) ([]domainusers.User, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return Service{repo: repo}
}

func (s Service) Save(ctx context.Context, user domainusers.User) error {
	if user.ID == "" {
		return fmt.Errorf("user id is required")
	}
	return s.repo.SaveUser(ctx, user)
}

func (s Service) PublicAuthor(ctx context.Context, id string) (domainusers.AuthorProfile, bool, error) {
	user, ok, err := s.repo.GetUser(ctx, id)
	if err != nil || !ok {
		return domainusers.AuthorProfile{}, ok, err
	}
	return user.PublicAuthor(), true, nil
}

func (s Service) List(ctx context.Context) ([]domainusers.AuthorProfile, error) {
	users, err := s.repo.ListUsers(ctx)
	if err != nil {
		return nil, err
	}
	authors := []domainusers.AuthorProfile{}
	for _, user := range users {
		if user.Active {
			authors = append(authors, user.PublicAuthor())
		}
	}
	return authors, nil
}
