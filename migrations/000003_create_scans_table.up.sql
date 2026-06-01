CREATE TABLE sekai.scans (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    owner_id BIGINT REFERENCES sekai.users(id) NOT NULL,
    status SMALLINT NOT NULL,
    api_type SMALLINT NOT NULL,

    artifact_schema_version INT NOT NULL,
    artifact_metadata JSONB NOT NULL,

    manifest_key TEXT DEFAULT '',
    container_manifest_key TEXT DEFAULT '',

    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);
