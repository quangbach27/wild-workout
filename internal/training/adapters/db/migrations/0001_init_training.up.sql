BEGIN;

CREATE SCHEMA IF NOT EXISTS training;

CREATE TABLE training.trainings (
    training_uuid uuid NOT NULL,
    user_uuid uuid NOT NULL,
    username varchar(255) NOT NULL,
    time TIMESTAMPTZ NOT NULL,
    notes text,
    proposed_time TIMESTAMPTZ,
    move_proposed_by varchar(255),

    canceled boolean NOT NULL,
    PRIMARY KEY (training_uuid)
);

COMMIT;