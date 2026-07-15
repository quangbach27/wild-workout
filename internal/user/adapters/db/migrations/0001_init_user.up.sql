BEGIN;

CREATE SCHEMA IF NOT EXISTS "user";

CREATE TABLE "user".users (
  user_uuid uuid PRIMARY KEY,
  firebase_uid varchar(255) UNIQUE NOT NULL,
  username varchar(255) NOT NULL,
  role varchar(100) NOT NULL,
  balance INT NOT NULL DEFAULT 5 CHECK (balance >= 0)
);

COMMIT;