CREATE TABLE IF NOT EXISTS %[1]s.refresh_tokens (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES %[1]s.users(id) ON DELETE CASCADE,
    hashed_refresh_token TEXT NOT NULL,
    access_token_jti TEXT NOT NULL,
    client_ip TEXT NOT NULL, 
    expires_at TIMESTAMP NOT NULL,
    CONSTRAINT unique_jti_per_user UNIQUE (user_id, access_token_jti)
);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON %[1]s.refresh_tokens (user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_jti ON %[1]s.refresh_tokens (access_token_jti);