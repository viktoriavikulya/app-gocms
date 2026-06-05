package settings

import (
	"context"
	"fmt"
	"time"

	domainsettings "github.com/fastygo/app-gocms/internal/domain/settings"
	"github.com/fastygo/app-gocms/internal/extensions"
)

type Repository interface {
	SaveSetting(context.Context, domainsettings.Value) error
	GetSetting(context.Context, string) (domainsettings.Value, bool, error)
	ListSettings(context.Context) ([]domainsettings.Value, error)
}

type Registry struct {
	definitions map[string]domainsettings.Definition
}

func NewRegistry(definitions ...domainsettings.Definition) Registry {
	registry := Registry{definitions: map[string]domainsettings.Definition{}}
	for _, definition := range definitions {
		registry.definitions[definition.Key] = definition
	}
	return registry
}

type Service struct {
	repo     Repository
	registry Registry
	hooks    *extensions.HookBus
	now      func() time.Time
}

func NewService(repo Repository, registry Registry) Service {
	return Service{repo: repo, registry: registry, now: time.Now}
}

func (s Service) WithHooks(bus *extensions.HookBus) Service {
	s.hooks = bus
	return s
}

func (s Service) Save(ctx context.Context, value domainsettings.Value) error {
	if value.Key == "" {
		return fmt.Errorf("setting key is required")
	}
	if definition, ok := s.registry.definitions[value.Key]; ok {
		value.Group = definition.Group
		value.Public = definition.Public
	}
	if err := s.dispatch(ctx, extensions.HookSettingsUpdateBefore, value); err != nil {
		return err
	}
	if err := s.repo.SaveSetting(ctx, value); err != nil {
		return err
	}
	_ = s.dispatch(ctx, extensions.HookSettingsUpdateAfter, value)
	return nil
}

func (s Service) dispatch(ctx context.Context, hook string, entity any) error {
	if s.hooks == nil {
		return nil
	}
	return s.hooks.Dispatch(ctx, hook, extensions.HookPayload{Hook: hook, Entity: entity, OccurredAt: s.now().UTC()})
}

func (s Service) Get(ctx context.Context, key string) (domainsettings.Value, bool, error) {
	value, ok, err := s.repo.GetSetting(ctx, key)
	if err != nil || ok {
		return value, ok, err
	}
	if definition, found := s.registry.definitions[key]; found {
		return domainsettings.Value{Key: key, Group: definition.Group, Value: definition.DefaultValue, Public: definition.Public}, true, nil
	}
	return domainsettings.Value{}, false, nil
}

func (s Service) Public(ctx context.Context) ([]domainsettings.Value, error) {
	values, err := s.repo.ListSettings(ctx)
	if err != nil {
		return nil, err
	}
	public := []domainsettings.Value{}
	for _, value := range values {
		if value.Public {
			value.Private = false
			public = append(public, value)
		}
	}
	return public, nil
}
