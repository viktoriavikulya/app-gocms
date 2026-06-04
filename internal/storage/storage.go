package storage

import (
	"context"

	"github.com/fastygo/platform/pkg/contracts"
)

type StoreProvider interface {
	ForWorkspace(contracts.WorkspaceID) WorkspaceStore
}

type WorkspaceStore interface {
	WithinTx(context.Context, func(context.Context, Repositories) error) error
}

type Repositories struct {
	Content ContentRepository
}

type ContentRepository interface {
	Put(context.Context, contracts.RecordID, contracts.Record) error
	List(context.Context, string) (contracts.PageResult, error)
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
		repos := Repositories{Content: contentRepository{workspace: s.workspace, tx: tx}}
		return fn(ctx, repos)
	})
}

type contentRepository struct {
	workspace contracts.WorkspaceID
	tx        contracts.StorageTx
}

func (r contentRepository) Put(ctx context.Context, id contracts.RecordID, record contracts.Record) error {
	return r.tx.Put(ctx, "content", id, record)
}

func (r contentRepository) List(ctx context.Context, kind string) (contracts.PageResult, error) {
	return r.tx.List(ctx, contracts.Query{Workspace: r.workspace, RecordType: kind})
}
