package contenttype

import (
	"context"
	"fmt"

	"github.com/fastygo/app-gocms/internal/domain/contenttype"
)

type Repository interface {
	SaveContentType(context.Context, contenttype.Type) error
	GetContentType(context.Context, contenttype.ID) (contenttype.Type, bool, error)
	ListContentTypes(context.Context) ([]contenttype.Type, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return Service{repo: repo}
}

func (s Service) InstallBuiltIns(ctx context.Context) error {
	for _, item := range []contenttype.Type{contenttype.BuiltInPost(), contenttype.BuiltInPage()} {
		if err := s.Register(ctx, item); err != nil {
			return err
		}
	}
	return nil
}

func (s Service) Register(ctx context.Context, item contenttype.Type) error {
	if item.ID == "" {
		return fmt.Errorf("content type id is required")
	}
	if item.Label == "" {
		return fmt.Errorf("content type label is required")
	}
	return s.repo.SaveContentType(ctx, item)
}

func (s Service) Get(ctx context.Context, id contenttype.ID) (contenttype.Type, bool, error) {
	return s.repo.GetContentType(ctx, id)
}

func (s Service) List(ctx context.Context) ([]contenttype.Type, error) {
	return s.repo.ListContentTypes(ctx)
}
