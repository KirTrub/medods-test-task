ALTER TABLE task_schedules ADD COLUMN deadline_days INT NOT NULL DEFAULT 0;

ALTER TABLE tasks ADD COLUMN deadline_at TIMESTAMPTZ;

CREATE INDEX idx_tasks_deadline_status ON tasks (deadline_at, status)
    WHERE status != 'done' AND status != 'overdue';

CREATE UNIQUE INDEX idx_tasks_schedule_today
    ON tasks (schedule_id, date_trunc('day', created_at))
    WHERE schedule_id IS NOT NULL;