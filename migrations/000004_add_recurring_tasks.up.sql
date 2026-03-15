ALTER TABLE note
ADD COLUMN due_at DATETIME;

ALTER TABLE note
ADD COLUMN task_series_id INTEGER REFERENCES task_series (id);

ALTER TABLE note
ADD COLUMN recurrence_rule TEXT;

ALTER TABLE note
ADD COLUMN recurrence_weekday INTEGER;

ALTER TABLE note
ADD COLUMN recurrence_day_of_month INTEGER;

CREATE TABLE
    task_series (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        content TEXT NOT NULL,
        recurrence_rule TEXT NOT NULL,
        recurrence_weekday INTEGER,
        recurrence_day_of_month INTEGER,
        active INTEGER NOT NULL DEFAULT 1,
        created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
        CHECK (recurrence_rule IN ('daily', 'weekly', 'monthly')),
        CHECK (
            recurrence_weekday IS NULL
            OR recurrence_weekday BETWEEN 0 AND 6
        ),
        CHECK (
            recurrence_day_of_month IS NULL
            OR recurrence_day_of_month BETWEEN 1 AND 31
        ),
        CHECK (active IN (0, 1))
    );

CREATE TABLE
    task_series_tags (
        task_series_id INTEGER NOT NULL REFERENCES task_series (id) ON DELETE CASCADE,
        tag_id INTEGER NOT NULL REFERENCES tags (id) ON DELETE CASCADE,
        PRIMARY KEY (task_series_id, tag_id)
    );

CREATE INDEX idx_note_task_series_id ON note (task_series_id);

CREATE INDEX idx_task_series_tags_series_id ON task_series_tags (task_series_id);

CREATE INDEX idx_task_series_tags_tag_id ON task_series_tags (tag_id);

CREATE UNIQUE INDEX idx_note_pending_task_series ON note (task_series_id)
WHERE
    task_series_id IS NOT NULL
    AND completed_at IS NULL;
