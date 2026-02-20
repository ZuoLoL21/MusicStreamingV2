CREATE TABLE bandit_data
(
    user_uuid UUID,
    theme Text,
    weights Text,
    biases Text,
    version int DEFAULT 0,
    updated_at timestamptz DEFAULT now(),
    PRIMARY KEY (user_uuid, theme)
);

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