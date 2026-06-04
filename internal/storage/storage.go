package storage

import (
	"context"

	"github.com/fastygo/app-gocms/pkg/module/records"
	"github.com/fastygo/platform/pkg/contracts"
)

type StoreProvider interface {
	ForWorkspace(contracts.WorkspaceID) WorkspaceStore
}

type WorkspaceStore interface {
	WithinTx(context.Context, func(context.Context, Repositories) error) error
}

type Repositories struct {
	Posts                  RecordRepository
	Pages                  RecordRepository
	ContentTypes           RecordRepository
	ContentMetaDefinitions RecordRepository
	Taxonomies             RecordRepository
	Terms                  RecordRepository
	MediaAssets            RecordRepository
	Authors                RecordRepository
	Settings               RecordRepository
	Menus                  RecordRepository
}

type RecordRepository interface {
	Put(context.Context, contracts.RecordID, contracts.Record) error
	Get(context.Context, contracts.RecordID) (contracts.Record, error)
	List(context.Context) (contracts.PageResult, error)
	Delete(context.Context, contracts.RecordID) error
}

type Provider struct {
	Port contracts.StoragePort
}

func NewProvider(port contracts.StoragePort) Provider {
	return Provider{Port: port}
}

func (p Provider) ForWorkspace(workspace contracts.WorkspaceID) WorkspaceStore {
	return workspaceStore{workspace: workspace, port: p.Port}
}

type workspaceStore struct {
	workspace contracts.WorkspaceID
	port      contracts.StoragePort
}

func (s workspaceStore) WithinTx(ctx context.Context, fn func(context.Context, Repositories) error) error {
	return s.port.WithinWorkspaceTx(ctx, s.workspace, func(ctx context.Context, tx contracts.StorageTx) error {
		repos := Repositories{
			Posts:                  recordRepository{workspace: s.workspace, recordType: string(records.RecordPost), tx: tx},
			Pages:                  recordRepository{workspace: s.workspace, recordType: string(records.RecordPage), tx: tx},
			ContentTypes:           recordRepository{workspace: s.workspace, recordType: string(records.RecordContentType), tx: tx},
			ContentMetaDefinitions: recordRepository{workspace: s.workspace, recordType: string(records.RecordContentMeta), tx: tx},
			Taxonomies:             recordRepository{workspace: s.workspace, recordType: string(records.RecordTaxonomy), tx: tx},
			Terms:                  recordRepository{workspace: s.workspace, recordType: string(records.RecordTerm), tx: tx},
			MediaAssets:            recordRepository{workspace: s.workspace, recordType: string(records.RecordMediaAsset), tx: tx},
			Authors:                recordRepository{workspace: s.workspace, recordType: string(records.RecordAuthor), tx: tx},
			Settings:               recordRepository{workspace: s.workspace, recordType: "setting", tx: tx},
			Menus:                  recordRepository{workspace: s.workspace, recordType: "menu", tx: tx},
		}
		return fn(ctx, repos)
	})
}

type recordRepository struct {
	workspace  contracts.WorkspaceID
	recordType string
	tx         contracts.StorageTx
}

func (r recordRepository) Put(ctx context.Context, id contracts.RecordID, record contracts.Record) error {
	return r.tx.Put(ctx, r.recordType, id, record)
}

func (r recordRepository) Get(ctx context.Context, id contracts.RecordID) (contracts.Record, error) {
	return r.tx.Get(ctx, r.recordType, id)
}

func (r recordRepository) List(ctx context.Context) (contracts.PageResult, error) {
	return r.tx.List(ctx, contracts.Query{Workspace: r.workspace, RecordType: r.recordType})
}

func (r recordRepository) Delete(ctx context.Context, id contracts.RecordID) error {
	return r.tx.Delete(ctx, r.recordType, id)
}
