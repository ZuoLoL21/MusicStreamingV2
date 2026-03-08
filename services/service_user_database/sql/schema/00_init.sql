-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Enable trigram extension for similarity search
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Create ENUM types
CREATE TYPE artist_member_role AS ENUM ('owner', 'manager', 'member');
