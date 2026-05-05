ALTER TABLE tasks
ADD COLUMN IF NOT EXISTS recurring_config JSONB,
ADD COLUMN IF NOT EXISTS parent_task_id BIGINT,
ADD COLUMN IF NOT EXISTS scheduled_date TIMESTAMPTZ,
ADD COLUMN IF NOT EXISTS effective_from TIMESTAMPTZ DEFAULT NOW();

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_tasks_parent') THEN
        ALTER TABLE tasks ADD CONSTRAINT fk_tasks_parent
        FOREIGN KEY (parent_task_id) REFERENCES tasks(id) ON DELETE CASCADE;
    END IF;
END $$;

ALTER TABLE tasks
ADD COLUMN IF NOT EXISTS rec_type TEXT GENERATED ALWAYS AS (recurring_config->>'type') STORED,
ADD COLUMN IF NOT EXISTS rec_day_of_month INT GENERATED ALWAYS AS ((recurring_config->>'day_of_month')::int) STORED;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_rec_type') THEN
        ALTER TABLE tasks ADD CONSTRAINT chk_rec_type
        CHECK (rec_type IS NULL OR rec_type IN ('daily', 'monthly', 'specific_dates', 'parity'));
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_rec_day_of_month') THEN
        ALTER TABLE tasks ADD CONSTRAINT chk_rec_day_of_month
        CHECK (rec_type IS NULL OR rec_type <> 'monthly' OR rec_day_of_month BETWEEN 1 AND 31);
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_rec_config_structure') THEN
        ALTER TABLE tasks ADD CONSTRAINT chk_rec_config_structure CHECK (
            recurring_config IS NULL OR
            CASE rec_type
                WHEN 'daily' THEN jsonb_typeof(recurring_config->'interval') = 'number'
                    AND (recurring_config->>'interval')::int >= 1
                WHEN 'monthly' THEN jsonb_typeof(recurring_config->'day_of_month') = 'number'
                    AND (recurring_config->>'day_of_month')::int BETWEEN 1 AND 31
                WHEN 'specific_dates' THEN jsonb_typeof(recurring_config->'dates') = 'array'
                    AND jsonb_array_length(recurring_config->'dates') > 0
                WHEN 'parity' THEN recurring_config->>'parity_mode' IN ('even', 'odd')
                ELSE FALSE
            END
        );
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'uq_parent_scheduled') THEN
        ALTER TABLE tasks ADD CONSTRAINT uq_parent_scheduled
        UNIQUE (parent_task_id, scheduled_date);
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_tasks_rec_type ON tasks(rec_type);
CREATE INDEX IF NOT EXISTS idx_tasks_parent_scheduled ON tasks(parent_task_id, scheduled_date) WHERE parent_task_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_tasks_active_templates ON tasks(rec_type, (recurring_config->>'start_date')) WHERE parent_task_id IS NULL AND recurring_config IS NOT NULL;