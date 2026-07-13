BEGIN;

CREATE SCHEMA IF NOT EXISTS trainer;

CREATE TYPE trainer.availability AS ENUM (
    'available',
    'not_available',
    'training_scheduled'
);

CREATE TABLE trainer.hours (
    hour         TIMESTAMPTZ NOT NULL,
    availability trainer.availability NOT NULL,
    PRIMARY KEY (hour)
);

COMMIT;