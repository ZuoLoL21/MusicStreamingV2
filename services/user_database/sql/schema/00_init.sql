-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create ENUM types
CREATE TYPE artist_member_role AS ENUM ('owner', 'manager', 'member');
