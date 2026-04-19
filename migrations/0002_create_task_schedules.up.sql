CREATE TABLE task_schedules (
    id           BIGSERIAL PRIMARY KEY,
    title        TEXT        NOT NULL,
    description  TEXT        NOT NULL DEFAULT '',
    type         TEXT        NOT NULL,
    every_n_days INT,
    day_of_month INT,
    dates        JSONB,
    parity       TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE tasks
    ADD COLUMN schedule_id BIGINT REFERENCES task_schedules(id) ON DELETE SET NULL;

CREATE INDEX idx_tasks_schedule_id ON tasks (schedule_id);