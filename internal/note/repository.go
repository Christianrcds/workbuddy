package note

import (
	"context"
	"database/sql"
)

type Repository interface {
	CreateNote(ctx context.Context, params CreateNoteParams) (Note, error)
	GetNote(ctx context.Context, id int64) (Note, error)
	DeleteNote(ctx context.Context, id int64) error
	ListNotes(ctx context.Context) ([]Note, error)
}

type sqliteRepo struct {
	queries *Queries
	db      *sql.DB
}

func (r *sqliteRepo) CreateNote(ctx context.Context, params CreateNoteParams) (Note, error) {
	return r.queries.CreateNote(ctx, params)
}

func (r *sqliteRepo) GetNote(ctx context.Context, id int64) (Note, error) {
	return r.queries.GetNote(ctx, id)
}

func (r *sqliteRepo) DeleteNote(ctx context.Context, id int64) error {
	return r.queries.DeleteNote(ctx, id)
}

func (r *sqliteRepo) ListNotes(ctx context.Context) ([]Note, error) {
	return r.queries.ListNotes(ctx)
}

func NewRepository(db *sql.DB) Repository {
	return &sqliteRepo{
		queries: New(db),
		db:      db,
	}
}
