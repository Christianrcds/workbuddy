-- Core tables
CREATE TABLE
    note (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        content TEXT NOT NULL,
        created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
    );

CREATE TABLE
    tags (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL UNIQUE,
        created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
    );

-- Many-to-many relationship between notes and tags
CREATE TABLE
    note_tags (
        note_id INTEGER NOT NULL REFERENCES note (id) ON DELETE CASCADE,
        tag_id INTEGER NOT NULL REFERENCES tags (id) ON DELETE CASCADE,
        PRIMARY KEY (note_id, tag_id)
    );

CREATE VIRTUAL TABLE notes_fts USING fts5 (
    name,
    content
);

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

CREATE INDEX idx_note_tags_note_id ON note_tags (note_id);

CREATE INDEX idx_note_tags_tag_id ON note_tags (tag_id);

CREATE INDEX idx_tags_name ON tags (name);

CREATE INDEX idx_note_created_at ON note (created_at);