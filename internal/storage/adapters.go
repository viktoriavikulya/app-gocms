package storage

import (
	"context"
	"encoding/json"
	"fmt"

	appcontent "github.com/fastygo/app-gocms/internal/application/content"
	appcontenttype "github.com/fastygo/app-gocms/internal/application/contenttype"
	appmedia "github.com/fastygo/app-gocms/internal/application/media"
	appmenus "github.com/fastygo/app-gocms/internal/application/menus"
	apppreview "github.com/fastygo/app-gocms/internal/application/preview"
	apprevisions "github.com/fastygo/app-gocms/internal/application/revisions"
	appsettings "github.com/fastygo/app-gocms/internal/application/settings"
	apptaxonomy "github.com/fastygo/app-gocms/internal/application/taxonomy"
	appusers "github.com/fastygo/app-gocms/internal/application/users"
	domaincontent "github.com/fastygo/app-gocms/internal/domain/content"
	"github.com/fastygo/app-gocms/internal/domain/contenttype"
	"github.com/fastygo/app-gocms/internal/domain/media"
	"github.com/fastygo/app-gocms/internal/domain/menus"
	"github.com/fastygo/app-gocms/internal/domain/preview"
	"github.com/fastygo/app-gocms/internal/domain/revisions"
	"github.com/fastygo/app-gocms/internal/domain/settings"
	"github.com/fastygo/app-gocms/internal/domain/taxonomy"
	"github.com/fastygo/app-gocms/internal/domain/users"
	"github.com/fastygo/platform/pkg/contracts"
)

var (
	_ appcontent.Repository        = ApplicationRepositories{}
	_ appcontent.TypeRegistry      = ApplicationRepositories{}
	_ appcontenttype.Repository    = ApplicationRepositories{}
	_ appsettings.Repository       = ApplicationRepositories{}
	_ appusers.Repository          = ApplicationRepositories{}
	_ apptaxonomy.Repository       = ApplicationRepositories{}
	_ apptaxonomy.EntryRepository  = ApplicationRepositories{}
	_ appmedia.Repository          = ApplicationRepositories{}
	_ appmedia.EntryRepository     = ApplicationRepositories{}
	_ appmenus.Repository          = ApplicationRepositories{}
	_ apprevisions.Repository      = ApplicationRepositories{}
	_ apprevisions.EntryRepository = ApplicationRepositories{}
	_ apppreview.Repository        = ApplicationRepositories{}
	_ apppreview.EntryRepository   = ApplicationRepositories{}
)

type ApplicationRepositories struct {
	repos Repositories
}

func NewApplicationRepositories(repos Repositories) ApplicationRepositories {
	return ApplicationRepositories{repos: repos}
}

func (a ApplicationRepositories) Save(ctx context.Context, entry domaincontent.Entry) error {
	repo, err := a.contentRepo(entry.Kind)
	if err != nil {
		return err
	}
	return put(ctx, repo, contracts.RecordID(entry.ID), entry)
}

func (a ApplicationRepositories) Get(ctx context.Context, id domaincontent.ID) (domaincontent.Entry, bool, error) {
	for _, repo := range []RecordRepository{a.repos.Posts, a.repos.Pages} {
		entry, ok, err := get[domaincontent.Entry](ctx, repo, contracts.RecordID(id))
		if err == nil && ok {
			return entry, true, nil
		}
	}
	return domaincontent.Entry{}, false, nil
}

func (a ApplicationRepositories) List(ctx context.Context, query domaincontent.Query) ([]domaincontent.Entry, error) {
	repos := []RecordRepository{}
	if query.Kind == "" || query.Kind == domaincontent.KindPost {
		repos = append(repos, a.repos.Posts)
	}
	if query.Kind == "" || query.Kind == domaincontent.KindPage {
		repos = append(repos, a.repos.Pages)
	}
	var entries []domaincontent.Entry
	for _, repo := range repos {
		items, err := list[domaincontent.Entry](ctx, repo)
		if err != nil {
			return nil, err
		}
		for _, entry := range items {
			if query.Status == "" || entry.Status == query.Status {
				entries = append(entries, entry)
			}
		}
	}
	return entries, nil
}

func (a ApplicationRepositories) SaveContentType(ctx context.Context, item contenttype.Type) error {
	return put(ctx, a.repos.ContentTypes, contracts.RecordID(item.ID), item)
}

func (a ApplicationRepositories) GetContentType(ctx context.Context, id contenttype.ID) (contenttype.Type, bool, error) {
	return get[contenttype.Type](ctx, a.repos.ContentTypes, contracts.RecordID(id))
}

func (a ApplicationRepositories) ListContentTypes(ctx context.Context) ([]contenttype.Type, error) {
	return list[contenttype.Type](ctx, a.repos.ContentTypes)
}

func (a ApplicationRepositories) SaveSetting(ctx context.Context, item settings.Value) error {
	return put(ctx, a.repos.Settings, contracts.RecordID(item.Key), item)
}

func (a ApplicationRepositories) GetSetting(ctx context.Context, key string) (settings.Value, bool, error) {
	return get[settings.Value](ctx, a.repos.Settings, contracts.RecordID(key))
}

func (a ApplicationRepositories) ListSettings(ctx context.Context) ([]settings.Value, error) {
	return list[settings.Value](ctx, a.repos.Settings)
}

func (a ApplicationRepositories) SaveUser(ctx context.Context, user users.User) error {
	return put(ctx, a.repos.Authors, contracts.RecordID(user.ID), user)
}

func (a ApplicationRepositories) GetUser(ctx context.Context, id string) (users.User, bool, error) {
	return get[users.User](ctx, a.repos.Authors, contracts.RecordID(id))
}

func (a ApplicationRepositories) ListUsers(ctx context.Context) ([]users.User, error) {
	return list[users.User](ctx, a.repos.Authors)
}

func (a ApplicationRepositories) SaveDefinition(ctx context.Context, definition taxonomy.Definition) error {
	return put(ctx, a.repos.Taxonomies, contracts.RecordID(definition.Type), definition)
}

func (a ApplicationRepositories) GetDefinition(ctx context.Context, taxonomyType string) (taxonomy.Definition, bool, error) {
	return get[taxonomy.Definition](ctx, a.repos.Taxonomies, contracts.RecordID(taxonomyType))
}

func (a ApplicationRepositories) ListDefinitions(ctx context.Context) ([]taxonomy.Definition, error) {
	return list[taxonomy.Definition](ctx, a.repos.Taxonomies)
}

func (a ApplicationRepositories) GetTerm(ctx context.Context, id string) (taxonomy.Term, bool, error) {
	return get[taxonomy.Term](ctx, a.repos.Terms, contracts.RecordID(id))
}

func (a ApplicationRepositories) SaveTerm(ctx context.Context, term taxonomy.Term) error {
	return put(ctx, a.repos.Terms, contracts.RecordID(term.ID), term)
}

func (a ApplicationRepositories) ListTerms(ctx context.Context, taxonomyType string) ([]taxonomy.Term, error) {
	terms, err := list[taxonomy.Term](ctx, a.repos.Terms)
	if err != nil {
		return nil, err
	}
	result := []taxonomy.Term{}
	for _, term := range terms {
		if taxonomyType == "" || term.TaxonomyType == taxonomyType {
			result = append(result, term)
		}
	}
	return result, nil
}

func (a ApplicationRepositories) SaveAsset(ctx context.Context, asset media.Asset) error {
	return put(ctx, a.repos.MediaAssets, contracts.RecordID(asset.ID), asset)
}

func (a ApplicationRepositories) GetAsset(ctx context.Context, id string) (media.Asset, bool, error) {
	return get[media.Asset](ctx, a.repos.MediaAssets, contracts.RecordID(id))
}

func (a ApplicationRepositories) ListAssets(ctx context.Context) ([]media.Asset, error) {
	return list[media.Asset](ctx, a.repos.MediaAssets)
}

func (a ApplicationRepositories) SaveMenu(ctx context.Context, menu menus.Menu) error {
	return put(ctx, a.repos.Menus, contracts.RecordID(menu.ID), menu)
}

func (a ApplicationRepositories) ListMenus(ctx context.Context) ([]menus.Menu, error) {
	return list[menus.Menu](ctx, a.repos.Menus)
}

func (a ApplicationRepositories) GetMenuByLocation(ctx context.Context, location string) (menus.Menu, bool, error) {
	items, err := a.ListMenus(ctx)
	if err != nil {
		return menus.Menu{}, false, err
	}
	for _, menu := range items {
		if menu.Location == location {
			return menu, true, nil
		}
	}
	return menus.Menu{}, false, nil
}

func (a ApplicationRepositories) SaveRevision(ctx context.Context, revision revisions.Revision) error {
	return put(ctx, a.repos.Revisions, contracts.RecordID(revision.ID), revision)
}

func (a ApplicationRepositories) GetRevision(ctx context.Context, id string) (revisions.Revision, bool, error) {
	return get[revisions.Revision](ctx, a.repos.Revisions, contracts.RecordID(id))
}

func (a ApplicationRepositories) SavePreview(ctx context.Context, access preview.Access) error {
	return put(ctx, a.repos.Previews, contracts.RecordID(access.Token), access)
}

func (a ApplicationRepositories) GetPreview(ctx context.Context, token string) (preview.Access, bool, error) {
	return get[preview.Access](ctx, a.repos.Previews, contracts.RecordID(token))
}

func (a ApplicationRepositories) contentRepo(kind domaincontent.Kind) (RecordRepository, error) {
	switch kind {
	case domaincontent.KindPost:
		return a.repos.Posts, nil
	case domaincontent.KindPage:
		return a.repos.Pages, nil
	default:
		return nil, fmt.Errorf("unsupported content kind %q", kind)
	}
}

func put(ctx context.Context, repo RecordRepository, id contracts.RecordID, value any) error {
	record, err := toRecord(value)
	if err != nil {
		return err
	}
	return repo.Put(ctx, id, record)
}

func get[T any](ctx context.Context, repo RecordRepository, id contracts.RecordID) (T, bool, error) {
	var zero T
	record, err := repo.Get(ctx, id)
	if err != nil {
		return zero, false, nil
	}
	value, err := fromRecord[T](record)
	if err != nil {
		return zero, false, err
	}
	return value, true, nil
}

func list[T any](ctx context.Context, repo RecordRepository) ([]T, error) {
	page, err := repo.List(ctx)
	if err != nil {
		return nil, err
	}
	result := []T{}
	for _, record := range page.Records {
		value, err := fromRecord[T](record)
		if err != nil {
			return nil, err
		}
		result = append(result, value)
	}
	return result, nil
}

func toRecord(value any) (contracts.Record, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	record := contracts.Record{}
	return record, json.Unmarshal(payload, &record)
}

func fromRecord[T any](record contracts.Record) (T, error) {
	var value T
	payload, err := json.Marshal(record)
	if err != nil {
		return value, err
	}
	return value, json.Unmarshal(payload, &value)
}
