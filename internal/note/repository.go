package note

import (
	"context"
	"database/sql"
)

type Repository interface {
	CreateNote(ctx context.Context, params string) (Note, error)
	ListNotes(ctx context.Context) ([]Note, error)

	// Tags
	CreateTag(ctx context.Context, name string) (Tag, error)
	GetTag(ctx context.Context, name string) (Tag, error)
	ListTags(ctx context.Context, name string) ([]Tag, error)

	// Note Tags Relationships
	AddTagToNote(ctx context.Context, arg AddTagToNoteParams) error
	GetNotesByTagWithLimit(ctx context.Context, arg GetNotesByTagWithLimitParams) ([]Note, error)
}

type sqliteRepo struct {
	queries *Queries
	db      *sql.DB
}

func (r *sqliteRepo) CreateNote(ctx context.Context, params string) (Note, error) {
	return r.queries.CreateNote(ctx, params)
}

func (r *sqliteRepo) ListNotes(ctx context.Context) ([]Note, error) {
	return r.queries.ListNotes(ctx)
}

func (r *sqliteRepo) CreateTag(ctx context.Context, name string) (Tag, error) {
	return r.queries.CreateTag(ctx, name)
}

func (r *sqliteRepo) GetTag(ctx context.Context, name string) (Tag, error) {
	return r.queries.GetTag(ctx, name)
}

func (r *sqliteRepo) AddTagToNote(ctx context.Context, arg AddTagToNoteParams) error {
	return r.queries.AddTagToNote(ctx, arg)
}

func (r *sqliteRepo) GetNotesByTagWithLimit(ctx context.Context, arg GetNotesByTagWithLimitParams) ([]Note, error) {
	return r.queries.GetNotesByTagWithLimit(ctx, arg)
}

func (r *sqliteRepo) ListTags(ctx context.Context, name string) ([]Tag, error) {
	return r.queries.ListTags(ctx, name)
}

func NewRepository(db *sql.DB) Repository {
	return &sqliteRepo{
		queries: New(db),
		db:      db,
	}
}

func NewRepositoryWithTx(tx *sql.Tx) Repository {
	return &sqliteRepo{
		queries: New(tx),
		db:      nil,
	}
}
