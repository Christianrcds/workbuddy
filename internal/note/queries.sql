-- Notes
-- name: CreateNote :one
INSERT INTO
    note (name, content)
VALUES
    (?, ?) RETURNING *;

-- name: GetNote :one
SELECT
    *
FROM
    note
WHERE
    id = ?;

-- name: ListNotes :many
SELECT
    *
FROM
    note
ORDER BY
    created_at DESC;

-- name: UpdateNote :one
UPDATE note
SET
    name = ?,
    content = ?
WHERE
    id = ? RETURNING *;

-- name: DeleteNote :exec
DELETE FROM note
WHERE
    id = ?;

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

-- name: ListTags :many
SELECT
    *
FROM
    tags
ORDER BY
    name;

-- name: DeleteTag :exec
DELETE FROM tags
WHERE
    id = ?;

-- Note Tags Relationships
-- name: AddTagToNote :exec
INSERT INTO
    note_tags (note_id, tag_id)
VALUES
    (?, ?);

-- name: RemoveTagFromNote :exec
DELETE FROM note_tags
WHERE
    note_id = ?
    AND tag_id = ?;

-- name: GetNoteTags :many
SELECT
    t.*
FROM
    tags t
    INNER JOIN note_tags nt ON t.id = nt.tag_id
WHERE
    nt.note_id = ?
ORDER BY
    t.name;

-- name: GetNotesByTag :many
SELECT
    n.*
FROM
    note n
    INNER JOIN note_tags nt ON n.id = nt.note_id
    INNER JOIN tags t ON nt.tag_id = t.id
WHERE
    t.name = ?
ORDER BY
    n.created_at DESC;

-- name: GetNotesWithTags :many
SELECT
    n.id,
    n.name,
    n.content,
    n.created_at,
    GROUP_CONCAT (t.name, ', ') as tag_names
FROM
    note n
    LEFT JOIN note_tags nt ON n.id = nt.note_id
    LEFT JOIN tags t ON nt.tag_id = t.id
GROUP BY
    n.id,
    n.name,
    n.content,
    n.created_at
ORDER BY
    n.created_at DESC;

-- Simple text search for now (FTS5 will be implemented manually)
-- name: SearchNotesSimple :many
SELECT * FROM note 
WHERE name LIKE ? OR content LIKE ?
ORDER BY created_at DESC;