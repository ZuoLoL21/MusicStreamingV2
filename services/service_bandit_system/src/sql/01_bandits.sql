CREATE TABLE IF NOT EXISTS bandit_data (
    user_uuid UUID NOT NULL,
    theme VARCHAR(255) NOT NULL,
    version INTEGER NOT NULL DEFAULT 0,
    weights JSONB NOT NULL,  -- 12x12 matrix serialized as JSON
    biases JSONB NOT NULL,   -- 12-element vector serialized as JSON
    weights_inv JSONB NOT NULL,  -- 12x12 inverse matrix serialized as JSON
    updates_since_recompute INTEGER DEFAULT 0,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_uuid, theme)
);
CREATE INDEX IF NOT EXISTS idx_bandit_user_theme ON bandit_data(user_uuid, theme);
CREATE INDEX IF NOT EXISTS idx_bandit_version ON bandit_data(user_uuid, theme, version);

-- Create function to update the updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Add triggers to tables with updated_at columns
CREATE TRIGGER update_bandit_data_updated_at
    BEFORE UPDATE ON bandit_data
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();