CREATE TABLE IF NOT EXISTS processed_files (
                                               id              BIGSERIAL PRIMARY KEY,
                                               file_name       TEXT        NOT NULL UNIQUE,
                                               status          TEXT        NOT NULL DEFAULT 'pending',
                                               error_detail    TEXT        NOT NULL DEFAULT '',
                                               processed_at    TIMESTAMPTZ,
                                               created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_processed_files_status ON processed_files (status);