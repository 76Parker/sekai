CREATE TABLE IF NOT EXISTS sekai.artifacts (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id INT REFERENCES sekai.users(id),
    name TEXT NOT NULL,
    sloc INT NOT NULL,
    archive_sha256 TEXT NOT NULL,
    project_sha256 TEXT NOT NULL,

    archive_size_bytes BIGINT NOT NULL,
    unpacked_size_bytes BIGINT NOT NULL,

    CONSTRAINT artifact_for_per_user_unique UNIQUE (user_id, project_sha256)
);
