CREATE TABLE IF NOT EXISTS sast_scan_source_functions (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    scan_id BIGINT NOT NULL REFERENCES sast_scans(id) ON DELETE CASCADE,
    source_function_id BIGINT NOT NULL REFERENCES sast_source_functions(id) ON DELETE CASCADE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (scan_id, source_function_id)
);
