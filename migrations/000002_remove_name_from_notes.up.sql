-- Drop existing triggers that reference the name column
DROP TRIGGER IF EXISTS notes_fts_insert;

DROP TRIGGER IF EXISTS notes_fts_delete;

DROP TRIGGER IF EXISTS notes_fts_update;

-- Drop the FTS table since it references name
DROP TABLE IF EXISTS notes_fts;

-- Remove name column from notes table
ALTER TABLE note
DROP COLUMN name;

-- Recreate FTS table without the name column
CREATE VIRTUAL TABLE notes_fts USING fts5 (content);

-- Recreate triggers for the new FTS table (without name)
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

CREATE TRIGGER notes_fts_update AFTER
UPDATE ON note BEGIN
DELETE FROM notes_fts
WHERE
    rowid = old.id;

INSERT INTO
    notes_fts (rowid, content)
VALUES
    (new.id, new.content);

END;