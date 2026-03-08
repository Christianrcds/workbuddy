package note

import (
	"context"
	"database/sql"
	"strings"
	"unicode"
)

type Service struct {
	db   *sql.DB
	repo Repository
}

type SearchParams struct {
	Query     string
	Tag       string
	Limit     int64
	TasksOnly bool
	Completed *bool // nil = no filter, &true = completed only, &false = pending only
}

func (s *Service) SearchNotes(ctx context.Context, params SearchParams) ([]Note, error) {
	isTaskFilter := int64(0)
	completedFilter := int64(0)
	pendingFilter := int64(0)
	if params.TasksOnly || params.Completed != nil {
		isTaskFilter = 1
	}
	if params.Completed != nil {
		if *params.Completed {
			completedFilter = 1
		} else {
			pendingFilter = 1
		}
	}

	query := buildContentQuery(params.Query)
	if query != "" {
		if params.Tag != "" {
			return s.repo.SearchNotesByTagAndContent(ctx, SearchNotesByTagAndContentParams{
				Content: query,
				Name:    params.Tag,
				Column3: isTaskFilter,
				Column4: completedFilter,
				Column5: pendingFilter,
				Limit:   params.Limit,
			})
		}

		return s.repo.SearchNotesByContent(ctx, SearchNotesByContentParams{
			Content: query,
			Column2: isTaskFilter,
			Column3: completedFilter,
			Column4: pendingFilter,
			Limit:   params.Limit,
		})
	}

	if params.Tag != "" {
		return s.repo.SearchNotesByTag(ctx, SearchNotesByTagParams{
			Name:    params.Tag,
			Column2: isTaskFilter,
			Column3: completedFilter,
			Column4: pendingFilter,
			Limit:   params.Limit,
		})
	}

	return s.repo.SearchNotes(ctx, SearchNotesParams{
		Column1: isTaskFilter,
		Column2: completedFilter,
		Column3: pendingFilter,
		Limit:   params.Limit,
	})
}

func (s *Service) CreateNote(ctx context.Context, content string, tags []string, isTask int64) (Note, error) {
	trx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Note{}, err
	}
	defer trx.Rollback()

	repo := NewRepositoryWithTx(trx)

	createdNote, err := repo.CreateNote(ctx, CreateNoteParams{Content: content, IsTask: isTask})
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

func buildContentQuery(query string) string {
	tokens := strings.FieldsFunc(query, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
	if len(tokens) == 0 {
		return ""
	}

	terms := make([]string, 0, len(tokens))
	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		terms = append(terms, strings.ToLower(token)+"*")
	}
	if len(terms) == 0 {
		return ""
	}

	return strings.Join(terms, " ")
}

func NewService(db *sql.DB) *Service {
	return &Service{
		db:   db,
		repo: NewRepository(db),
	}
}
