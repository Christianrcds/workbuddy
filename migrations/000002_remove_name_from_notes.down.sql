DROP TRIGGER IF EXISTS notes_fts_insert;

DROP TRIGGER IF EXISTS notes_fts_delete;

DROP TRIGGER IF EXISTS notes_fts_update;

DROP TABLE IF EXISTS notes_fts;

ALTER TABLE note
ADD COLUMN name TEXT NOT NULL DEFAULT '';

CREATE VIRTUAL TABLE notes_fts USING fts5 (name, content);

CREATE TRIGGER notes_fts_insert AFTER INSERT ON note BEGIN
INSERT INTO
    notes_fts (rowid, name, content)
VALUES
    (new.id, new.name, new.content);

END;

CREATE TRIGGER notes_fts_delete AFTER DELETE ON note BEGIN
DELETE FROM notes_fts
WHERE
    rowid = old.id;

END;

CREATE TRIGGER notes_fts_update AFTER
UPDATE ON note BEGIN
DELETE FROM notes_fts
WHERE
    rowid = old.id;

INSERT INTO
    notes_fts (rowid, name, content)
VALUES
    (new.id, new.name, new.content);

END;