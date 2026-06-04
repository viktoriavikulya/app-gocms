package menus

import (
	"context"
	"fmt"

	domainmenus "github.com/fastygo/app-gocms/internal/domain/menus"
)

type Repository interface {
	SaveMenu(context.Context, domainmenus.Menu) error
	ListMenus(context.Context) ([]domainmenus.Menu, error)
	GetMenuByLocation(context.Context, string) (domainmenus.Menu, bool, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return Service{repo: repo}
}

func (s Service) Save(ctx context.Context, menu domainmenus.Menu) error {
	if menu.ID == "" || menu.Location == "" {
		return fmt.Errorf("menu id and location are required")
	}
	return s.repo.SaveMenu(ctx, menu)
}

func (s Service) List(ctx context.Context) ([]domainmenus.Menu, error) {
	return s.repo.ListMenus(ctx)
}

func (s Service) ByLocation(ctx context.Context, location string) (domainmenus.Menu, bool, error) {
	return s.repo.GetMenuByLocation(ctx, location)
}
