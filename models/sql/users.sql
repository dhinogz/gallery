CREATE TABLE user (
    id SERIAL PRIMARY KEY
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL
);