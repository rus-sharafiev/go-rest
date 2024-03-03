CREATE TYPE access AS ENUM ('USER', 'ADMIN');
CREATE TABLE users (
    "id" SERIAL PRIMARY KEY,
    "email" text UNIQUE,
    "firstName" text,
    "lastName" text,
    "phone" text UNIQUE,
    "avatar" text,
    "access" access DEFAULT 'USER',
    "createdAt" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP WITH TIME ZONE
);
CREATE TABLE passwords (
    "userId" integer REFERENCES users (id) ON DELETE CASCADE,
    "passwordHash" text
);
CREATE TABLE sessions (
    "userId" integer REFERENCES users (id) ON DELETE CASCADE,
    "fingerprint" text UNIQUE,
    "userAgent" text,
    "ip" text,
    "updatedAt" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);