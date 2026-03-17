CREATE TABLE IF NOT EXISTS refresh_tokens (
    uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_uuid UUID NOT NULL REFERENCES users(uuid) ON DELETE CASCADE,
    device_id UUID NOT NULL,
    token_hash VARCHAR(64) NOT NULL,
    device_name VARCHAR(255),
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_uuid, device_id)
);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_uuid ON refresh_tokens(user_uuid);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
