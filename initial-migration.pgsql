
CREATE TABLE users (
    id UUID PRIMARY KEY,
    provider VARCHAR(50) NOT NULL,
    provider_user_id VARCHAR(255) UNIQUE NOT NULL,
    nickname VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) NULL,
    avatar_url TEXT,
    access_token TEXT,
    refresh_token TEXT,
    provider_access_token TEXT,
    provider_refresh_token TEXT,
    status INT NOT NULL,
    "role" INT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NULL,
    terms_accepted BOOLEAN NOT NULL
);


DELETE FROM users;