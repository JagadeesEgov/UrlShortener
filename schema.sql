-- Clean URL Shortener Schema
-- This replaces all previous schema files

-- Drop existing table and sequence if they exist
DROP TABLE IF EXISTS eg_url_shortener CASCADE;
DROP SEQUENCE IF EXISTS eg_url_shorter_id;

-- Create sequence for auto-incrementing IDs
CREATE SEQUENCE eg_url_shorter_id START 1;

-- Create the URL shortener table with correct schema
CREATE TABLE eg_url_shortener (
    id BIGINT PRIMARY KEY DEFAULT nextval('eg_url_shorter_id'),
    url VARCHAR(2048) NOT NULL,
    validfrom BIGINT,
    validto BIGINT
);

-- Create index on URL for faster lookups
CREATE INDEX idx_eg_url_shortener_url ON eg_url_shortener(url);

-- Create index on validity dates for cleanup operations
CREATE INDEX idx_eg_url_shortener_validity ON eg_url_shortener(validfrom, validto);

-- Add comments for documentation
COMMENT ON TABLE eg_url_shortener IS 'URL shortener table for storing shortened URLs';
COMMENT ON COLUMN eg_url_shortener.id IS 'Auto-incrementing primary key';
COMMENT ON COLUMN eg_url_shortener.url IS 'Original long URL';
COMMENT ON COLUMN eg_url_shortener.validfrom IS 'Start time in milliseconds (Unix timestamp)';
COMMENT ON COLUMN eg_url_shortener.validto IS 'End time in milliseconds (Unix timestamp)'; 