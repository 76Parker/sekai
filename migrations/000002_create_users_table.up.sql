CREATE TABLE sekai.users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    username TEXT NOT NULL,
    email TEXT,
    scan_quota INT DEFAULT 10,
);
