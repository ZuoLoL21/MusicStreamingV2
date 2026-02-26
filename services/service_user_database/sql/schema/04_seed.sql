
-- Insert Sample Users
INSERT INTO users (uuid, username, email, hashed_password, bio) VALUES
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'default', 'john@example.com', '$2a$10$hashedpassword1', 'Very real human man');

-- Insert Genre Tags
INSERT INTO music_tags (tag_name, tag_description) VALUES
    ('Rock', 'Rock music genre'),
    ('Pop', 'Pop music genre'),
    ('Jazz', 'Jazz music genre'),
    ('Electronic', 'Electronic music genre'),
    ('Classical', 'Classical music genre'),
    ('Hip Hop', 'Hip Hop music genre'),
    ('R&B', 'R&B music genre'),
    ('Country', 'Country music genre'),
    ('Metal', 'Metal music genre'),
    ('Blues', 'Blues music genre'),
    ('Folk', 'Folk music genre'),
    ('Indie', 'Indie music genre'),
    ('Alternative', 'Alternative music genre'),
    ('Dance', 'Dance music genre'),
    ('Reggae', 'Reggae music genre');
