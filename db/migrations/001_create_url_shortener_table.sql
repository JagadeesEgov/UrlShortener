CREATE TABLE IF NOT EXISTS eg_url_shortener (
    short_key VARCHAR(4) PRIMARY KEY,
    url VARCHAR(2048) NOT NULL UNIQUE,
    validfrom BIGINT,
    validto BIGINT
);

