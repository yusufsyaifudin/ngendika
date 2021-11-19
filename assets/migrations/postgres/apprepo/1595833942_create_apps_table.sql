-- +migrate Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE IF NOT EXISTS apps (
    id UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
    client_id VARCHAR NOT NULL,
    name VARCHAR NOT NULL DEFAULT '',
    enabled BOOL NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

-- unique constraint prevented at code level
CREATE INDEX idx_apps_client_id ON apps (LOWER(client_id));

-- +migrate Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE IF EXISTS apps;
