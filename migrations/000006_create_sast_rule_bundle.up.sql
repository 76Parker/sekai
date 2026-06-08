CREATE TABLE IF NOT EXISTS sast_rule_bundle (
    id BIGINT GENERATED ALWAYS AS IDENTITY NOT NULL,
    tool SMALLINT NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    version TEXT NOT NULL,
    s3_key TEXT NOT NULL,
    s3_bucket TEXT NOT NULL,
)
