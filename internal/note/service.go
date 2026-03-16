package note

import (
	"context"
	"database/sql"
	"errors"
	"slices"
	"strings"
	"time"
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

type RecurrenceRule string

const (
	RecurrenceDaily   RecurrenceRule = "daily"
	RecurrenceWeekly  RecurrenceRule = "weekly"
	RecurrenceMonthly RecurrenceRule = "monthly"
)

type CreateParams struct {
	Content              string
	Tags                 []string
	IsTask               bool
	DueDate              *time.Time
	RecurrenceRule       RecurrenceRule
	RecurrenceWeekday    *time.Weekday
	RecurrenceDayOfMonth *int
}

type UpdateParams struct {
	Content              *string
	SetTags              []string
	AddTags              []string
	RemoveTags           []string
	DueDate              *time.Time
	ClearDueDate         bool
	RecurrenceRule       *RecurrenceRule
	RecurrenceWeekday    *time.Weekday
	RecurrenceDayOfMonth *int
	ClearRecurrence      bool
}

type CompletionResult struct {
	Completed Note
	Next      *Note
}

type RemoveResult struct {
	RemovedSeries    bool
	PreservedHistory bool
}

var (
	errEmptyNoteContent            = errors.New("note content cannot be empty")
	errDueDateRequiresTask         = errors.New("due dates are only available for tasks")
	errRecurringRequiresTask       = errors.New("recurrence is only available for tasks")
	errRecurringRequiresDueDate    = errors.New("recurring tasks require --due")
	errInvalidRecurrenceRule       = errors.New("recurrence must be one of: daily, weekly, monthly")
	errWeeklyRequiresWeekday       = errors.New("weekly recurrence requires --weekday")
	errMonthlyRequiresDayOfMonth   = errors.New("monthly recurrence requires --day-of-month")
	errDailyRejectsExtraSchedule   = errors.New("daily recurrence does not accept --weekday or --day-of-month")
	errWeeklyRejectsDayOfMonth     = errors.New("weekly recurrence does not accept --day-of-month")
	errMonthlyRejectsWeekday       = errors.New("monthly recurrence does not accept --weekday")
	errDueDateWeekdayMismatch      = errors.New("due date must match the selected recurrence weekday")
	errDueDateDayOfMonthMismatch   = errors.New("due date must match the selected recurrence day of month")
	errRecurrenceRequiresRule      = errors.New("recurrence options require --recur")
	errConflictingDueDateUpdate    = errors.New("cannot combine --due and --clear-due")
	errConflictingRecurrenceUpdate = errors.New("cannot combine --recur and --clear-recur")
	errCannotClearDueOnRecurring   = errors.New("cannot clear the due date of a recurring task without clearing recurrence")
	errTaskNotCompletable          = errors.New("task not found, already completed, or not a task")
)

func ErrTaskNotCompletable() error {
	return errTaskNotCompletable
}

type resolvedSchedule struct {
	dueDate     *time.Time
	rule        RecurrenceRule
	weekday     *time.Weekday
	dayOfMonth  *int
	isRecurring bool
}

type tagLookupCreator interface {
	CreateTag(ctx context.Context, name string) (Tag, error)
	GetTag(ctx context.Context, name string) (Tag, error)
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
	return s.CreateNoteWithParams(ctx, CreateParams{
		Content: content,
		Tags:    tags,
		IsTask:  isTask == 1,
	})
}

func (s *Service) CreateNoteWithParams(ctx context.Context, params CreateParams) (Note, error) {
	content := strings.TrimSpace(params.Content)
	if content == "" {
		return Note{}, errEmptyNoteContent
	}

	schedule, err := resolveSchedule(params.IsTask, params.DueDate, params.RecurrenceRule, params.RecurrenceWeekday, params.RecurrenceDayOfMonth)
	if err != nil {
		return Note{}, err
	}

	trx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Note{}, err
	}
	defer trx.Rollback()

	q := New(trx)
	normalizedTags := normalizeTags(params.Tags)

	var series TaskSeries
	seriesID := sql.NullInt64{}
	if schedule.isRecurring {
		series, err = q.CreateTaskSeries(ctx, CreateTaskSeriesParams{
			Content:              content,
			RecurrenceRule:       string(schedule.rule),
			RecurrenceWeekday:    schedule.weekdayNull(),
			RecurrenceDayOfMonth: schedule.dayOfMonthNull(),
		})
		if err != nil {
			return Note{}, err
		}
		seriesID = sql.NullInt64{Int64: series.ID, Valid: true}
		if err := syncTaskSeriesTags(ctx, q, series.ID, normalizedTags); err != nil {
			return Note{}, err
		}
	}

	createdNote, err := q.CreateNote(ctx, CreateNoteParams{
		Content:              content,
		IsTask:               boolToInt64(params.IsTask),
		DueAt:                schedule.dueDateNull(),
		TaskSeriesID:         seriesID,
		RecurrenceRule:       schedule.ruleNull(),
		RecurrenceWeekday:    schedule.weekdayNull(),
		RecurrenceDayOfMonth: schedule.dayOfMonthNull(),
	})
	if err != nil {
		return Note{}, err
	}

	if err := syncNoteTags(ctx, q, createdNote.ID, normalizedTags); err != nil {
		return Note{}, err
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

	q := New(trx)
	existing, err := q.GetNoteByID(ctx, id)
	if err != nil {
		return Note{}, err
	}

	if existing.TaskSeriesID.Valid {
		series, err := q.GetTaskSeriesByID(ctx, existing.TaskSeriesID.Int64)
		if err == nil && series.Active == 1 {
			updated, err := s.updateRecurringNote(ctx, q, existing, series, params)
			if err != nil {
				return Note{}, err
			}
			if err := trx.Commit(); err != nil {
				return Note{}, err
			}
			return updated, nil
		}
	}

	updated, err := s.updateStandaloneNote(ctx, q, existing, params)
	if err != nil {
		return Note{}, err
	}

	if err := trx.Commit(); err != nil {
		return Note{}, err
	}

	return updated, nil
}

func (s *Service) CompleteTask(ctx context.Context, id int64) (CompletionResult, error) {
	trx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return CompletionResult{}, err
	}
	defer trx.Rollback()

	q := New(trx)
	current, err := q.GetNoteByID(ctx, id)
	if err != nil {
		return CompletionResult{}, err
	}
	if current.IsTask != 1 {
		return CompletionResult{}, errTaskNotCompletable
	}

	rows, err := q.MarkNoteCompleted(ctx, id)
	if err != nil {
		return CompletionResult{}, err
	}
	if rows == 0 {
		return CompletionResult{}, errTaskNotCompletable
	}

	completed, err := q.GetNoteByID(ctx, id)
	if err != nil {
		return CompletionResult{}, err
	}

	result := CompletionResult{Completed: completed}

	if current.TaskSeriesID.Valid {
		series, err := q.GetTaskSeriesByID(ctx, current.TaskSeriesID.Int64)
		if err != nil {
			return CompletionResult{}, err
		}
		if series.Active == 1 {
			if !current.DueAt.Valid {
				return CompletionResult{}, errRecurringRequiresDueDate
			}

			nextDueDate, err := nextDueDate(current.DueAt.Time, series.RecurrenceRule, series.RecurrenceWeekday, series.RecurrenceDayOfMonth)
			if err != nil {
				return CompletionResult{}, err
			}
			tags, err := q.ListTagsByTaskSeriesID(ctx, series.ID)
			if err != nil {
				return CompletionResult{}, err
			}

			next, err := q.CreateNote(ctx, CreateNoteParams{
				Content:              series.Content,
				IsTask:               1,
				DueAt:                sql.NullTime{Time: nextDueDate, Valid: true},
				TaskSeriesID:         sql.NullInt64{Int64: series.ID, Valid: true},
				RecurrenceRule:       sql.NullString{String: series.RecurrenceRule, Valid: true},
				RecurrenceWeekday:    series.RecurrenceWeekday,
				RecurrenceDayOfMonth: series.RecurrenceDayOfMonth,
			})
			if err != nil {
				return CompletionResult{}, err
			}
			if err := syncNoteTags(ctx, q, next.ID, tags); err != nil {
				return CompletionResult{}, err
			}
			result.Next = &next
		}
	}

	if err := trx.Commit(); err != nil {
		return CompletionResult{}, err
	}

	return result, nil
}

func (s *Service) RemoveNote(ctx context.Context, id int64) (RemoveResult, error) {
	trx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return RemoveResult{}, err
	}
	defer trx.Rollback()

	q := New(trx)
	current, err := q.GetNoteByID(ctx, id)
	if err != nil {
		return RemoveResult{}, err
	}

	result := RemoveResult{}
	if current.TaskSeriesID.Valid {
		if _, err := q.SetTaskSeriesInactive(ctx, current.TaskSeriesID.Int64); err != nil {
			return RemoveResult{}, err
		}
		if _, err := q.DeletePendingNotesByTaskSeriesID(ctx, current.TaskSeriesID); err != nil {
			return RemoveResult{}, err
		}
		result.RemovedSeries = true
		result.PreservedHistory = true
	} else {
		rows, err := q.DeleteNoteByID(ctx, id)
		if err != nil {
			return RemoveResult{}, err
		}
		if rows == 0 {
			return RemoveResult{}, sql.ErrNoRows
		}
	}

	if err := trx.Commit(); err != nil {
		return RemoveResult{}, err
	}

	return result, nil
}

func (s *Service) updateStandaloneNote(ctx context.Context, q *Queries, existing Note, params UpdateParams) (Note, error) {
	content, err := resolveUpdatedContent(existing.Content, params.Content)
	if err != nil {
		return Note{}, err
	}
	if err := validateUpdateFlags(params); err != nil {
		return Note{}, err
	}

	currentTags, err := q.ListTagsByNoteID(ctx, existing.ID)
	if err != nil {
		return Note{}, err
	}
	desiredTags := buildDesiredTags(currentTags, params)

	baseSchedule := scheduleFromNote(existing)
	schedule, err := applyUpdateSchedule(existing.IsTask == 1, baseSchedule, params)
	if err != nil {
		return Note{}, err
	}

	seriesID := existing.TaskSeriesID
	if schedule.isRecurring {
		var series TaskSeries
		if existing.TaskSeriesID.Valid {
			series, err = q.GetTaskSeriesByID(ctx, existing.TaskSeriesID.Int64)
			if err != nil {
				return Note{}, err
			}
			if _, err := q.UpdateTaskSeries(ctx, UpdateTaskSeriesParams{
				Content:              content,
				RecurrenceRule:       string(schedule.rule),
				RecurrenceWeekday:    schedule.weekdayNull(),
				RecurrenceDayOfMonth: schedule.dayOfMonthNull(),
				Active:               1,
				ID:                   series.ID,
			}); err != nil {
				return Note{}, err
			}
		} else {
			series, err = q.CreateTaskSeries(ctx, CreateTaskSeriesParams{
				Content:              content,
				RecurrenceRule:       string(schedule.rule),
				RecurrenceWeekday:    schedule.weekdayNull(),
				RecurrenceDayOfMonth: schedule.dayOfMonthNull(),
			})
			if err != nil {
				return Note{}, err
			}
			seriesID = sql.NullInt64{Int64: series.ID, Valid: true}
		}
		if err := syncTaskSeriesTags(ctx, q, seriesID.Int64, desiredTags); err != nil {
			return Note{}, err
		}
	} else if existing.TaskSeriesID.Valid && params.ClearRecurrence {
		if _, err := q.SetTaskSeriesInactive(ctx, existing.TaskSeriesID.Int64); err != nil {
			return Note{}, err
		}
		seriesID = sql.NullInt64{}
	}

	updated, err := q.UpdateNoteAttributesByID(ctx, UpdateNoteAttributesByIDParams{
		Content:              content,
		DueAt:                schedule.dueDateNull(),
		TaskSeriesID:         seriesID,
		RecurrenceRule:       schedule.ruleNull(),
		RecurrenceWeekday:    schedule.weekdayNull(),
		RecurrenceDayOfMonth: schedule.dayOfMonthNull(),
		ID:                   existing.ID,
	})
	if err != nil {
		return Note{}, err
	}

	if err := syncNoteTags(ctx, q, existing.ID, desiredTags); err != nil {
		return Note{}, err
	}

	return updated, nil
}

func (s *Service) updateRecurringNote(ctx context.Context, q *Queries, existing Note, series TaskSeries, params UpdateParams) (Note, error) {
	if err := validateUpdateFlags(params); err != nil {
		return Note{}, err
	}

	pending, err := q.GetPendingNoteByTaskSeriesID(ctx, existing.TaskSeriesID)
	if err == sql.ErrNoRows && !existing.CompletedAt.Valid {
		pending = existing
		err = nil
	}
	if err != nil {
		return Note{}, err
	}

	content, err := resolveUpdatedContent(series.Content, params.Content)
	if err != nil {
		return Note{}, err
	}

	currentTags, err := q.ListTagsByTaskSeriesID(ctx, series.ID)
	if err != nil {
		return Note{}, err
	}
	desiredTags := buildDesiredTags(currentTags, params)

	baseSchedule := scheduleFromSeries(series, pending.DueAt)
	schedule, err := applyUpdateSchedule(true, baseSchedule, params)
	if err != nil {
		return Note{}, err
	}

	if params.ClearRecurrence {
		if _, err := q.SetTaskSeriesInactive(ctx, series.ID); err != nil {
			return Note{}, err
		}
		if pending.ID != 0 {
			pending, err = q.UpdateNoteAttributesByID(ctx, UpdateNoteAttributesByIDParams{
				Content:              content,
				DueAt:                schedule.dueDateNull(),
				TaskSeriesID:         sql.NullInt64{},
				RecurrenceRule:       sql.NullString{},
				RecurrenceWeekday:    sql.NullInt64{},
				RecurrenceDayOfMonth: sql.NullInt64{},
				ID:                   pending.ID,
			})
			if err != nil {
				return Note{}, err
			}
			if err := syncNoteTags(ctx, q, pending.ID, desiredTags); err != nil {
				return Note{}, err
			}
		}
		if pending.ID == existing.ID {
			return pending, nil
		}
		return existing, nil
	}

	if _, err := q.UpdateTaskSeries(ctx, UpdateTaskSeriesParams{
		Content:              content,
		RecurrenceRule:       string(schedule.rule),
		RecurrenceWeekday:    schedule.weekdayNull(),
		RecurrenceDayOfMonth: schedule.dayOfMonthNull(),
		Active:               1,
		ID:                   series.ID,
	}); err != nil {
		return Note{}, err
	}
	if err := syncTaskSeriesTags(ctx, q, series.ID, desiredTags); err != nil {
		return Note{}, err
	}

	if pending.ID != 0 {
		pending, err = q.UpdateNoteAttributesByID(ctx, UpdateNoteAttributesByIDParams{
			Content:              content,
			DueAt:                schedule.dueDateNull(),
			TaskSeriesID:         sql.NullInt64{Int64: series.ID, Valid: true},
			RecurrenceRule:       schedule.ruleNull(),
			RecurrenceWeekday:    schedule.weekdayNull(),
			RecurrenceDayOfMonth: schedule.dayOfMonthNull(),
			ID:                   pending.ID,
		})
		if err != nil {
			return Note{}, err
		}
		if err := syncNoteTags(ctx, q, pending.ID, desiredTags); err != nil {
			return Note{}, err
		}
	}

	if pending.ID == existing.ID {
		return pending, nil
	}
	return existing, nil
}

func resolveUpdatedContent(existing string, incoming *string) (string, error) {
	if incoming == nil {
		return existing, nil
	}

	content := strings.TrimSpace(*incoming)
	if content == "" {
		return "", errEmptyNoteContent
	}
	return content, nil
}

func validateUpdateFlags(params UpdateParams) error {
	if params.ClearDueDate && params.DueDate != nil {
		return errConflictingDueDateUpdate
	}
	if params.ClearRecurrence && params.RecurrenceRule != nil {
		return errConflictingRecurrenceUpdate
	}
	return nil
}

func applyUpdateSchedule(isTask bool, base resolvedSchedule, params UpdateParams) (resolvedSchedule, error) {
	dueDate := copyTime(base.dueDate)
	rule := base.rule
	weekday := copyWeekday(base.weekday)
	dayOfMonth := copyInt(base.dayOfMonth)

	if params.ClearRecurrence {
		rule = ""
		weekday = nil
		dayOfMonth = nil
	}

	if params.RecurrenceRule != nil {
		rule = normalizeRecurrenceRule(*params.RecurrenceRule)
		switch rule {
		case RecurrenceDaily:
			weekday = nil
			dayOfMonth = nil
		case RecurrenceWeekly:
			dayOfMonth = nil
		case RecurrenceMonthly:
			weekday = nil
		}
	}

	if params.RecurrenceWeekday != nil {
		weekday = copyWeekday(params.RecurrenceWeekday)
	}
	if params.RecurrenceDayOfMonth != nil {
		dayOfMonth = copyInt(params.RecurrenceDayOfMonth)
	}

	if params.ClearDueDate {
		if rule != "" {
			return resolvedSchedule{}, errCannotClearDueOnRecurring
		}
		dueDate = nil
	}
	if params.DueDate != nil {
		dueDate = copyTime(params.DueDate)
	}

	return resolveSchedule(isTask, dueDate, rule, weekday, dayOfMonth)
}

func resolveSchedule(isTask bool, dueDate *time.Time, recurrenceRule RecurrenceRule, recurrenceWeekday *time.Weekday, recurrenceDayOfMonth *int) (resolvedSchedule, error) {
	rule := normalizeRecurrenceRule(recurrenceRule)

	if rule == "" {
		if dueDate != nil && !isTask {
			return resolvedSchedule{}, errDueDateRequiresTask
		}
		if recurrenceWeekday != nil || recurrenceDayOfMonth != nil {
			return resolvedSchedule{}, errRecurrenceRequiresRule
		}
		return resolvedSchedule{
			dueDate: copyTime(dueDate),
		}, nil
	}
	if !isTask {
		return resolvedSchedule{}, errRecurringRequiresTask
	}
	if dueDate == nil {
		return resolvedSchedule{}, errRecurringRequiresDueDate
	}

	switch rule {
	case RecurrenceDaily:
		if recurrenceWeekday != nil || recurrenceDayOfMonth != nil {
			return resolvedSchedule{}, errDailyRejectsExtraSchedule
		}
	case RecurrenceWeekly:
		if recurrenceWeekday == nil {
			return resolvedSchedule{}, errWeeklyRequiresWeekday
		}
		if recurrenceDayOfMonth != nil {
			return resolvedSchedule{}, errWeeklyRejectsDayOfMonth
		}
		if dueDate.Weekday() != *recurrenceWeekday {
			return resolvedSchedule{}, errDueDateWeekdayMismatch
		}
	case RecurrenceMonthly:
		if recurrenceDayOfMonth == nil {
			return resolvedSchedule{}, errMonthlyRequiresDayOfMonth
		}
		if recurrenceWeekday != nil {
			return resolvedSchedule{}, errMonthlyRejectsWeekday
		}
		if dueDate.Day() != *recurrenceDayOfMonth {
			return resolvedSchedule{}, errDueDateDayOfMonthMismatch
		}
	default:
		return resolvedSchedule{}, errInvalidRecurrenceRule
	}

	return resolvedSchedule{
		dueDate:     copyTime(dueDate),
		rule:        rule,
		weekday:     copyWeekday(recurrenceWeekday),
		dayOfMonth:  copyInt(recurrenceDayOfMonth),
		isRecurring: true,
	}, nil
}

func scheduleFromNote(note Note) resolvedSchedule {
	return resolvedSchedule{
		dueDate:     copyTime(nullTimePtr(note.DueAt)),
		rule:        normalizeRecurrenceRule(RecurrenceRule(note.RecurrenceRule.String)),
		weekday:     intToWeekday(note.RecurrenceWeekday),
		dayOfMonth:  intFromNull(note.RecurrenceDayOfMonth),
		isRecurring: note.RecurrenceRule.Valid,
	}
}

func scheduleFromSeries(series TaskSeries, dueAt sql.NullTime) resolvedSchedule {
	return resolvedSchedule{
		dueDate:     copyTime(nullTimePtr(dueAt)),
		rule:        normalizeRecurrenceRule(RecurrenceRule(series.RecurrenceRule)),
		weekday:     intToWeekday(series.RecurrenceWeekday),
		dayOfMonth:  intFromNull(series.RecurrenceDayOfMonth),
		isRecurring: series.Active == 1,
	}
}

func nextDueDate(current time.Time, rule string, weekday sql.NullInt64, dayOfMonth sql.NullInt64) (time.Time, error) {
	switch normalizeRecurrenceRule(RecurrenceRule(rule)) {
	case RecurrenceDaily:
		return current.AddDate(0, 0, 1), nil
	case RecurrenceWeekly:
		return current.AddDate(0, 0, 7), nil
	case RecurrenceMonthly:
		targetDay := current.Day()
		if dayOfMonth.Valid {
			targetDay = int(dayOfMonth.Int64)
		}
		year, month, _ := current.Date()
		nextMonth := time.Date(year, month, 1, current.Hour(), current.Minute(), current.Second(), current.Nanosecond(), current.Location()).AddDate(0, 1, 0)
		lastDay := daysInMonth(nextMonth.Year(), nextMonth.Month())
		if targetDay > lastDay {
			targetDay = lastDay
		}
		return time.Date(nextMonth.Year(), nextMonth.Month(), targetDay, current.Hour(), current.Minute(), current.Second(), current.Nanosecond(), current.Location()), nil
	default:
		return time.Time{}, errInvalidRecurrenceRule
	}
}

func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func syncNoteTags(ctx context.Context, q *Queries, noteID int64, tags []string) error {
	if err := q.DeleteAllTagsFromNote(ctx, noteID); err != nil {
		return err
	}
	for _, tagName := range normalizeTags(tags) {
		tag, err := getOrCreateTag(ctx, q, tagName)
		if err != nil {
			return err
		}
		if err := q.AddTagToNote(ctx, AddTagToNoteParams{
			NoteID: noteID,
			TagID:  tag.ID,
		}); err != nil {
			return err
		}
	}
	return nil
}

func syncTaskSeriesTags(ctx context.Context, q *Queries, taskSeriesID int64, tags []string) error {
	if err := q.DeleteAllTagsFromTaskSeries(ctx, taskSeriesID); err != nil {
		return err
	}
	for _, tagName := range normalizeTags(tags) {
		tag, err := getOrCreateTag(ctx, q, tagName)
		if err != nil {
			return err
		}
		if err := q.AddTagToTaskSeries(ctx, AddTagToTaskSeriesParams{
			TaskSeriesID: taskSeriesID,
			TagID:        tag.ID,
		}); err != nil {
			return err
		}
	}
	return nil
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

func getOrCreateTag(ctx context.Context, store tagLookupCreator, tagName string) (Tag, error) {
	tag, err := store.GetTag(ctx, tagName)
	if err == sql.ErrNoRows {
		return store.CreateTag(ctx, tagName)
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
	slices.Sort(normalized)
	return normalized
}

func normalizeTag(tag string) string {
	return strings.TrimSpace(tag)
}

func normalizeRecurrenceRule(rule RecurrenceRule) RecurrenceRule {
	return RecurrenceRule(strings.ToLower(strings.TrimSpace(string(rule))))
}

func boolToInt64(v bool) int64 {
	if v {
		return 1
	}
	return 0
}

func copyTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func copyWeekday(value *time.Weekday) *time.Weekday {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func copyInt(value *int) *int {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func nullTimePtr(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	cloned := value.Time
	return &cloned
}

func intToWeekday(value sql.NullInt64) *time.Weekday {
	if !value.Valid {
		return nil
	}
	weekday := time.Weekday(value.Int64)
	return &weekday
}

func intFromNull(value sql.NullInt64) *int {
	if !value.Valid {
		return nil
	}
	cloned := int(value.Int64)
	return &cloned
}

func (s resolvedSchedule) dueDateNull() sql.NullTime {
	if s.dueDate == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *s.dueDate, Valid: true}
}

func (s resolvedSchedule) ruleNull() sql.NullString {
	if s.rule == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: string(s.rule), Valid: true}
}

func (s resolvedSchedule) weekdayNull() sql.NullInt64 {
	if s.weekday == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(*s.weekday), Valid: true}
}

func (s resolvedSchedule) dayOfMonthNull() sql.NullInt64 {
	if s.dayOfMonth == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(*s.dayOfMonth), Valid: true}
}

func NewService(db *sql.DB) *Service {
	return &Service{
		db:   db,
		repo: NewRepository(db),
	}
}
