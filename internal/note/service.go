package note

import (
	"context"
	"database/sql"
	"errors"
	"slices"
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

var errEmptyNoteContent = errors.New("note content cannot be empty")

type UpdateParams struct {
	Content    *string
	SetTags    []string
	AddTags    []string
	RemoveTags []string
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

func (s *Service) GetNoteByID(ctx context.Context, id int64) (Note, error) {
	return s.repo.GetNoteByID(ctx, id)
}

func (s *Service) UpdateNoteContent(ctx context.Context, id int64, content string) (Note, error) {
	return s.UpdateNote(ctx, id, UpdateParams{
		Content: &content,
	})
}

func (s *Service) UpdateNote(ctx context.Context, id int64, params UpdateParams) (Note, error) {
	trx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Note{}, err
	}
	defer trx.Rollback()

	repo := NewRepositoryWithTx(trx)

	updatedNote, err := repo.GetNoteByID(ctx, id)
	if err != nil {
		return Note{}, err
	}

	if params.Content != nil {
		content := strings.TrimSpace(*params.Content)
		if content == "" {
			return Note{}, errEmptyNoteContent
		}

		updatedNote, err = repo.UpdateNoteContentByID(ctx, UpdateNoteContentByIDParams{
			Content: content,
			ID:      id,
		})
		if err != nil {
			return Note{}, err
		}
	}

	currentTags, err := repo.ListTagsByNoteID(ctx, id)
	if err != nil {
		return Note{}, err
	}

	desiredTags := buildDesiredTags(currentTags, params)
	if !slicesEqual(currentTags, desiredTags) {
		if err := repo.DeleteAllTagsFromNote(ctx, id); err != nil {
			return Note{}, err
		}
		for _, tagName := range desiredTags {
			tag, err := getOrCreateTag(ctx, repo, tagName)
			if err != nil {
				return Note{}, err
			}
			if err := repo.AddTagToNote(ctx, AddTagToNoteParams{
				NoteID: id,
				TagID:  tag.ID,
			}); err != nil {
				return Note{}, err
			}
		}
	}

	if err := trx.Commit(); err != nil {
		return Note{}, err
	}

	return updatedNote, nil
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

func buildDesiredTags(currentTags []string, params UpdateParams) []string {
	desired := make(map[string]struct{}, len(currentTags))
	for _, tag := range currentTags {
		normalized := normalizeTag(tag)
		if normalized == "" {
			continue
		}
		desired[normalized] = struct{}{}
	}

	if len(params.SetTags) > 0 {
		desired = make(map[string]struct{})
		for _, tag := range normalizeTags(params.SetTags) {
			desired[tag] = struct{}{}
		}
	}

	for _, tag := range normalizeTags(params.AddTags) {
		desired[tag] = struct{}{}
	}

	for _, tag := range normalizeTags(params.RemoveTags) {
		delete(desired, tag)
	}

	out := make([]string, 0, len(desired))
	for tag := range desired {
		out = append(out, tag)
	}
	slices.Sort(out)
	return out
}

func getOrCreateTag(ctx context.Context, repo Repository, tagName string) (Tag, error) {
	tag, err := repo.GetTag(ctx, tagName)
	if err == sql.ErrNoRows {
		return repo.CreateTag(ctx, tagName)
	}
	return tag, err
}

func normalizeTags(tags []string) []string {
	seen := make(map[string]struct{}, len(tags))
	normalized := make([]string, 0, len(tags))
	for _, tag := range tags {
		name := normalizeTag(tag)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		normalized = append(normalized, name)
	}
	return normalized
}

func normalizeTag(tag string) string {
	return strings.TrimSpace(tag)
}

func slicesEqual(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func NewService(db *sql.DB) *Service {
	return &Service{
		db:   db,
		repo: NewRepository(db),
	}
}
