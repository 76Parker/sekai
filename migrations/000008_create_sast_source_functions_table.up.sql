CREATE TABLE IF NOT EXISTS sast_source_functions (
    id BIGINT GENERATED ALWAYS AS IDENTITY NOT NULL,
    scan_id BIGINT REFERENCES sast_scans(id) ON DELETE CASCADE,
    hash TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    signature TEXT NOT NULL,
    kind TEXT NOT NULL,
    package_name TEXT NOT NULL,
    package_path TEXT NOT NULL,
    file TEXT NOT NULL,
    start_line INT NOT NULL,
    end_line INT NOT NULL,
    source TEXT NOT NULL
)
