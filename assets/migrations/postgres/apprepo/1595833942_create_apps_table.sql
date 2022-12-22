-- +migrate Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE IF NOT EXISTS apps (
    id BIGINT NOT NULL PRIMARY KEY,
    client_id VARCHAR NOT NULL,
    name VARCHAR NOT NULL DEFAULT '',

    -- using unix microsecond to make it easier to migrate between db
    created_at BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM now()) * 1000000),
    updated_at BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM now()) * 1000000),

    -- ensure that this only one record that not deleted
    deleted_at BIGINT NOT NULL DEFAULT 0
);

-- unique constraint prevented at sql level
-- https://dba.stackexchange.com/a/9760
-- client_id | deleted_at
-- a         | 2022-10-10 10:04:00 -> OK
-- a         | 2022-10-10 10:05:00 -> OK
-- a         | 1970-01-01 00:00:00 -> OK
-- a         | 1970-01-01 00:00:00 -> FAILED because previous data has NULL 1970-01-01 00:00:00
-- We're not using NULL since it need more handling!
-- We assume that record with or below date 1970-01-01 00:00:00 is valid, and greater than that date is deleted (tombstone).
-- Try use: INSERT INTO apps(id, client_id, name, enabled) (SELECT x, concat('app', x::text), concat('app', x::text), true FROM generate_series(1,10000) AS x);
-- Fast query: SELECT * FROM apps WHERE LOWER(client_id) = 'app8843' AND deleted_at <= to_timestamp(0) ORDER BY id DESC LIMIT 1;
CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_apps_client_id_deleted ON apps (LOWER(client_id), deleted_at);

-- +migrate Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE IF EXISTS apps;
