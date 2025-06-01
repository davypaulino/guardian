
CREATE TABLE users (
    id UUID PRIMARY KEY,
    provider VARCHAR(50),
    provider_user_id VARCHAR(255) UNIQUE,
    nickname VARCHAR(255),
    email VARCHAR(255) UNIQUE NULL,
    avatar_url TEXT,
    access_token TEXT,
    refresh_token TEXT,
    provider_access_token TEXT,
    provider_refresh_token TEXT,
    status INT,
    "role" INT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NULL
);

DROP TABLE users CASCADE;
