DROP INDEX IF EXISTS idx_note_pending_task_series;
DROP INDEX IF EXISTS idx_task_series_tags_tag_id;
DROP INDEX IF EXISTS idx_task_series_tags_series_id;
DROP INDEX IF EXISTS idx_note_task_series_id;

DROP TABLE IF EXISTS task_series_tags;
DROP TABLE IF EXISTS task_series;

CREATE TABLE note_recurring_down (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    content TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME,
    is_task INTEGER NOT NULL DEFAULT 0
);

INSERT INTO
    note_recurring_down (id, content, created_at, completed_at, is_task)
SELECT
    id,
    content,
    created_at,
    completed_at,
    is_task
FROM
    note;

DROP TRIGGER IF EXISTS notes_fts_insert;
DROP TRIGGER IF EXISTS notes_fts_delete;
DROP TRIGGER IF EXISTS notes_fts_update;
DROP TABLE IF EXISTS notes_fts;

DROP TABLE note;
ALTER TABLE note_recurring_down RENAME TO note;

CREATE VIRTUAL TABLE notes_fts USING fts5 (content);

INSERT INTO
    notes_fts (rowid, content)
SELECT
    id,
    content
FROM
    note;

CREATE TRIGGER notes_fts_insert AFTER INSERT ON note BEGIN
INSERT INTO
    notes_fts (rowid, content)
VALUES
    (new.id, new.content);
END;

CREATE TRIGGER notes_fts_delete AFTER DELETE ON note BEGIN
DELETE FROM notes_fts
WHERE
    rowid = old.id;
END;

CREATE TRIGGER notes_fts_update AFTER UPDATE ON note BEGIN
DELETE FROM notes_fts
WHERE
    rowid = old.id;

INSERT INTO
    notes_fts (rowid, content)
VALUES
    (new.id, new.content);
END;
