CREATE TABLE IF NOT EXISTS sast_findings (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    check_id BIGINT REFERENCES sast_checks(id),
    rule_id BIGINT REFERENCES sast_rules(id),
    fingerprint BYTEA,
    primary_location jsonb,
    flow jsonb
)
