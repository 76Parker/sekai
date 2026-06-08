CREATE TABLE IF NOT EXISTS sast_rules (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    bundle_id BIGINT REFERENCES sast_rule_bundle(id) NOT NULL,
    query_path TEXT NOT NULL,
    name TEXT NOT NULL,
    CWE SMALLINT NOT NULL,
    severity SMALLINT NOT NULL,
    description TEXT NOT NULL,
);
