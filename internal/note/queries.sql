-- Notes
-- name: CreateNote :one
INSERT INTO
  note (
    content,
    is_task,
    due_at,
    task_series_id,
    recurrence_rule,
    recurrence_weekday,
    recurrence_day_of_month
  )
VALUES
  (?, ?, ?, ?, ?, ?, ?) RETURNING id,
  content,
  created_at,
  completed_at,
  is_task,
  due_at,
  task_series_id,
  recurrence_rule,
  recurrence_weekday,
  recurrence_day_of_month;

-- name: ListNotes :many
SELECT
  id,
  content,
  created_at,
  completed_at,
  is_task,
  due_at,
  task_series_id,
  recurrence_rule,
  recurrence_weekday,
  recurrence_day_of_month
FROM
  note
ORDER BY
  created_at DESC;

-- name: ListTasks :many
SELECT
  id,
  content,
  created_at,
  completed_at,
  is_task,
  due_at,
  task_series_id,
  recurrence_rule,
  recurrence_weekday,
  recurrence_day_of_month
FROM
  note
WHERE
  is_task = 1
ORDER BY
  created_at DESC;

-- name: GetNoteByID :one
SELECT
  id,
  content,
  created_at,
  completed_at,
  is_task,
  due_at,
  task_series_id,
  recurrence_rule,
  recurrence_weekday,
  recurrence_day_of_month
FROM
  note
WHERE
  id = ?;

-- name: GetPendingNoteByTaskSeriesID :one
SELECT
  id,
  content,
  created_at,
  completed_at,
  is_task,
  due_at,
  task_series_id,
  recurrence_rule,
  recurrence_weekday,
  recurrence_day_of_month
FROM
  note
WHERE
  task_series_id = ?
  AND completed_at IS NULL;

-- name: UpdateNoteAttributesByID :one
UPDATE note
SET
  content = ?,
  due_at = ?,
  task_series_id = ?,
  recurrence_rule = ?,
  recurrence_weekday = ?,
  recurrence_day_of_month = ?
WHERE
  id = ? RETURNING id,
  content,
  created_at,
  completed_at,
  is_task,
  due_at,
  task_series_id,
  recurrence_rule,
  recurrence_weekday,
  recurrence_day_of_month;

-- name: DeletePendingNotesByTaskSeriesID :execrows
DELETE FROM note
WHERE
  task_series_id = ?
  AND completed_at IS NULL;

-- Tags
-- name: CreateTag :one
INSERT INTO
  tags (name)
VALUES
  (?) RETURNING *;

-- name: GetTag :one
SELECT
  *
FROM
  tags
WHERE
  name = ?;

-- Note Tags Relationships
-- name: AddTagToNote :exec
INSERT INTO
  note_tags (note_id, tag_id)
VALUES
  (?, ?);

-- name: DeleteAllTagsFromNote :exec
DELETE FROM note_tags
WHERE
  note_id = ?;

-- Task series
-- name: CreateTaskSeries :one
INSERT INTO
  task_series (
    content,
    recurrence_rule,
    recurrence_weekday,
    recurrence_day_of_month
  )
VALUES
  (?, ?, ?, ?) RETURNING id,
  content,
  recurrence_rule,
  recurrence_weekday,
  recurrence_day_of_month,
  active,
  created_at;

-- name: GetTaskSeriesByID :one
SELECT
  id,
  content,
  recurrence_rule,
  recurrence_weekday,
  recurrence_day_of_month,
  active,
  created_at
FROM
  task_series
WHERE
  id = ?;

-- name: GetTaskSeriesByNoteID :one
SELECT
  ts.id,
  ts.content,
  ts.recurrence_rule,
  ts.recurrence_weekday,
  ts.recurrence_day_of_month,
  ts.active,
  ts.created_at
FROM
  task_series ts
  INNER JOIN note n ON n.task_series_id = ts.id
WHERE
  n.id = ?;

-- name: UpdateTaskSeries :one
UPDATE task_series
SET
  content = ?,
  recurrence_rule = ?,
  recurrence_weekday = ?,
  recurrence_day_of_month = ?,
  active = ?
WHERE
  id = ? RETURNING id,
  content,
  recurrence_rule,
  recurrence_weekday,
  recurrence_day_of_month,
  active,
  created_at;

-- name: SetTaskSeriesInactive :execrows
UPDATE task_series
SET
  active = 0
WHERE
  id = ?;

-- name: AddTagToTaskSeries :exec
INSERT INTO
  task_series_tags (task_series_id, tag_id)
VALUES
  (?, ?);

-- name: DeleteAllTagsFromTaskSeries :exec
DELETE FROM task_series_tags
WHERE
  task_series_id = ?;

-- name: ListTagsByTaskSeriesID :many
SELECT
  t.name
FROM
  tags t
  INNER JOIN task_series_tags tst ON t.id = tst.tag_id
WHERE
  tst.task_series_id = ?
ORDER BY
  t.name;

-- name: SearchNotesByTag :many 
SELECT
  n.id,
  n.content,
  n.created_at,
  n.completed_at,
  n.is_task,
  n.due_at,
  n.task_series_id,
  n.recurrence_rule,
  n.recurrence_weekday,
  n.recurrence_day_of_month
FROM
  note n
  INNER JOIN note_tags nt ON n.id = nt.note_id
  INNER JOIN tags t ON nt.tag_id = t.id
WHERE
  t.name = ?
  AND (
    ? = 0
    OR n.is_task = 1
  )
  AND (
    ? = 0
    OR n.completed_at IS NOT NULL
  )
  AND (
    ? = 0
    OR n.completed_at IS NULL
  )
ORDER BY
  n.created_at DESC
LIMIT
  ?;

-- name: SearchNotes :many
SELECT
  id,
  content,
  created_at,
  completed_at,
  is_task,
  due_at,
  task_series_id,
  recurrence_rule,
  recurrence_weekday,
  recurrence_day_of_month
FROM
  note
WHERE
  (
    ? = 0
    OR is_task = 1
  )
  AND (
    ? = 0
    OR completed_at IS NOT NULL
  )
  AND (
    ? = 0
    OR completed_at IS NULL
  )
ORDER BY
  created_at DESC
LIMIT
  ?;

-- name: SearchNotesByContent :many
SELECT
  n.id,
  n.content,
  n.created_at,
  n.completed_at,
  n.is_task,
  n.due_at,
  n.task_series_id,
  n.recurrence_rule,
  n.recurrence_weekday,
  n.recurrence_day_of_month
FROM
  notes_fts
  INNER JOIN note n ON n.id = notes_fts.rowid
WHERE
  notes_fts.content MATCH ?
  AND (
    ? = 0
    OR n.is_task = 1
  )
  AND (
    ? = 0
    OR n.completed_at IS NOT NULL
  )
  AND (
    ? = 0
    OR n.completed_at IS NULL
  )
ORDER BY
  n.created_at DESC
LIMIT
  ?;

-- name: SearchNotesByTagAndContent :many
SELECT
  n.id,
  n.content,
  n.created_at,
  n.completed_at,
  n.is_task,
  n.due_at,
  n.task_series_id,
  n.recurrence_rule,
  n.recurrence_weekday,
  n.recurrence_day_of_month
FROM
  notes_fts
  INNER JOIN note n ON n.id = notes_fts.rowid
  INNER JOIN note_tags nt ON n.id = nt.note_id
  INNER JOIN tags t ON nt.tag_id = t.id
WHERE
  notes_fts.content MATCH ?
  AND t.name = ?
  AND (
    ? = 0
    OR n.is_task = 1
  )
  AND (
    ? = 0
    OR n.completed_at IS NOT NULL
  )
  AND (
    ? = 0
    OR n.completed_at IS NULL
  )
ORDER BY
  n.created_at DESC
LIMIT
  ?;

-- name: MarkNoteCompleted :execrows
UPDATE note
SET
  completed_at = CURRENT_TIMESTAMP
WHERE
  id = ?
  AND completed_at IS NULL
  AND is_task = 1;

-- name: UpdateNoteContentByID :one
UPDATE note
SET
  content = ?
WHERE
  id = ? RETURNING id,
  content,
  created_at,
  completed_at,
  is_task,
  due_at,
  task_series_id,
  recurrence_rule,
  recurrence_weekday,
  recurrence_day_of_month;

-- name: DeleteTaskByID :execrows
DELETE FROM note
WHERE
  id = ?
  AND is_task = 1;

-- name: DeleteNoteByID :execrows
DELETE FROM note
WHERE
  id = ?;

-- name: ListTags :many
SELECT
  *
FROM
  tags
WHERE
  name LIKE ?;

-- name: ListTagsByNoteID :many
SELECT
  t.name
FROM
  tags t
  INNER JOIN note_tags nt ON t.id = nt.tag_id
WHERE
  nt.note_id = ?
ORDER BY
  t.name;

-- name: ListTagsByNoteIDs :many
SELECT
  nt.note_id,
  t.name
FROM
  note_tags nt
  INNER JOIN tags t ON t.id = nt.tag_id
WHERE
  nt.note_id IN (sqlc.slice('note_ids'))
ORDER BY
  nt.note_id,
  t.name;
