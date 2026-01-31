package note

import (
	"context"
	"database/sql"
	"strings"
)

type Service struct {
	db   *sql.DB
	repo Repository
}

func (s *Service) CreateNoteWithTags(ctx context.Context, content string, tags []string) (Note, error) {
	trx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Note{}, err
	}
	defer trx.Rollback()

	repo := NewRepositoryWithTx(trx)

	createdNote, err := repo.CreateNote(ctx, content)

	if err != nil {
		return Note{}, err
	}

	for _, tagName := range tags {
		tagName = strings.TrimSpace(tagName)

		tag, err := repo.GetTag(ctx, tagName)
		if err == sql.ErrNoRows {
			tag, err = repo.CreateTag(ctx, tagName)
			if err != nil {
				return Note{}, err
			}
		} else if err != nil {
			return Note{}, err
		}
		err = repo.AddTagToNote(ctx, AddTagToNoteParams{
			NoteID: createdNote.ID,
			TagID:  tag.ID,
		})
		if err != nil {
			return Note{}, err
		}
	}

	if err := trx.Commit(); err != nil {
		return Note{}, err
	}

	return createdNote, nil
}

func NewService(db *sql.DB) *Service {
	return &Service{
		db:   db,
		repo: NewRepository(db),
	}
}
