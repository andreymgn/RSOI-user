CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    uid UUID PRIMARY KEY,
    username VARCHAR(30) NOT NULL UNIQUE,
    password_hash CHAR(60) NOT NULL,
    is_admin BOOLEAN DEFAULT FALSE
);

CREATE TABLE apps (
    uid UUID PRIMARY KEY,
    secret UUID NOT NULL,
    owner UUID REFERENCES users (uid),
    name VARCHAR(30) NOT NULL
);