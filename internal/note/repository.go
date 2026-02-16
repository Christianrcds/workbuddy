package note

import (
	"context"
	"database/sql"
)

type Repository interface {
	CreateNote(ctx context.Context, params CreateNoteParams) (Note, error)
	DeleteNoteByID(ctx context.Context, id int64) (int64, error)
	DeleteTaskByID(ctx context.Context, id int64) (int64, error)
	ListNotes(ctx context.Context) ([]Note, error)
	ListTasks(ctx context.Context) ([]Note, error)
	MarkNoteCompleted(ctx context.Context, id int64) (int64, error)

	// Tags
	CreateTag(ctx context.Context, name string) (Tag, error)
	GetTag(ctx context.Context, name string) (Tag, error)
	ListTags(ctx context.Context, name string) ([]Tag, error)
	ListTagsByNoteID(ctx context.Context, noteID int64) ([]string, error)

	// Note Tags Relationships
	AddTagToNote(ctx context.Context, arg AddTagToNoteParams) error
	GetNotesByTagWithLimit(ctx context.Context, arg GetNotesByTagWithLimitParams) ([]Note, error)
}

type sqliteRepo struct {
	queries *Queries
	db      *sql.DB
}

func (r *sqliteRepo) CreateNote(ctx context.Context, params CreateNoteParams) (Note, error) {
	return r.queries.CreateNote(ctx, params)
}

func (r *sqliteRepo) DeleteNoteByID(ctx context.Context, id int64) (int64, error) {
	return r.queries.DeleteNoteByID(ctx, id)
}

func (r *sqliteRepo) DeleteTaskByID(ctx context.Context, id int64) (int64, error) {
	return r.queries.DeleteTaskByID(ctx, id)
}

func (r *sqliteRepo) ListNotes(ctx context.Context) ([]Note, error) {
	return r.queries.ListNotes(ctx)
}

func (r *sqliteRepo) MarkNoteCompleted(ctx context.Context, id int64) (int64, error) {
	return r.queries.MarkNoteCompleted(ctx, id)
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

func (r *sqliteRepo) ListTagsByNoteID(ctx context.Context, noteID int64) ([]string, error) {
	return r.queries.ListTagsByNoteID(ctx, noteID)
}

func (r *sqliteRepo) ListTasks(ctx context.Context) ([]Note, error) {
	return r.queries.ListTasks(ctx)
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
