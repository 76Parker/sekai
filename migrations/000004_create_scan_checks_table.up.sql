CREATE TABLE sekai.scan_checks (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    scan_id BIGINT REFERENCES sekai.scans(id) ON DELETE CASCADE NOT NULL,
    check_type TEXT NOT NULL,
    status SMALLINT NOT NULL,
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ
);
