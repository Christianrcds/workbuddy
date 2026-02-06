-- Notes
-- name: CreateNote :one
INSERT INTO
    note (content)
VALUES
    (?) RETURNING *;

-- name: ListNotes :many
SELECT
    *
FROM
    note
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
    n.*
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
    *
FROM
    note
WHERE
    content LIKE ?
ORDER BY
    created_at DESC;

-- name: ListTags :many
SELECT
    *
FROM
    tags
WHERE
    name LIKE ?;