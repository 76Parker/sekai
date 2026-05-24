CREATE TABLE sekai.scans (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    artifact_id BIGINT REFERENCES sekai.artifacts(id),
    status TEXT NOT NULL,
    type SMALLINT NOT NULL,
    started_at TIMESTAMP NOT NULL,
    finished_at TIMESTAMP,
    buildable BOOLEAN NOT NULL
);
