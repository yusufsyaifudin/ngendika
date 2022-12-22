-- +migrate Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE IF NOT EXISTS push_providers (
    id BIGINT NOT NULL,
    app_id BIGINT NOT NULL,
    provider VARCHAR NOT NULL, -- fcm, apns, email
    label VARCHAR NOT NULL, -- name of service provider config, i.e: app-driver, app-consumer
    credential_json JSONB NOT NULL DEFAULT '{}', -- credential based on service_provider type

    -- using unix microsecond to make it easier to migrate between db
    created_at BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM now()) * 1000000),
    updated_at BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM now()) * 1000000),

    CONSTRAINT push_providers_pkey  PRIMARY KEY (id, app_id)
) PARTITION BY LIST (app_id);

CREATE INDEX idx_push_providers_provider_label ON ONLY push_providers (provider, label);


-- +migrate Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE IF EXISTS push_providers;
