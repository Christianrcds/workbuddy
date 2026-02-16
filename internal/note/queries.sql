-- Notes
-- name: CreateNote :one
INSERT INTO
    note (content, is_task)
VALUES
    (?, ?) RETURNING id, content, created_at, completed_at, is_task;

-- name: ListNotes :many
SELECT
    id,
    content,
    created_at,
    completed_at,
    is_task
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
    is_task
FROM
    note
WHERE
    is_task = 1
ORDER BY
    created_at DESC;

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

-- name: GetNotesByTagWithLimit :many
SELECT
    n.id,
    n.content,
    n.created_at,
    n.completed_at,
    n.is_task
FROM
    note n
    INNER JOIN note_tags nt ON n.id = nt.note_id
    INNER JOIN tags t ON nt.tag_id = t.id
WHERE
    t.name = ?
ORDER BY
    n.created_at DESC
LIMIT
    ?;

-- Simple text search for now (FTS5 will be implemented manually)
-- name: SearchNotesSimple :many
SELECT
    id,
    content,
    created_at,
    completed_at,
    is_task
FROM
    note
WHERE
    content LIKE ?
ORDER BY
    created_at DESC;

-- name: MarkNoteCompleted :execrows
UPDATE
    note
SET
    completed_at = CURRENT_TIMESTAMP
WHERE
    id = ?
    AND completed_at IS NULL
    AND is_task = 1;

-- name: DeleteTaskByID :execrows
DELETE FROM
    note
WHERE
    id = ?
    AND is_task = 1;

-- name: DeleteNoteByID :execrows
DELETE FROM
    note
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
