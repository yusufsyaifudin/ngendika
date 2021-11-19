-- +migrate Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE IF NOT EXISTS fcm_service_account_keys (
    id UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
    app_id UUID NOT NULL,
    service_account_key JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

-- for faster query WHERE app_client_id = ? ORDER BY created_at DESC
CREATE INDEX idx_fcm_service_account_keys_app_id_created_at ON fcm_service_account_keys (app_id, created_at DESC);

-- +migrate Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE IF EXISTS fcm_service_account_keys;
