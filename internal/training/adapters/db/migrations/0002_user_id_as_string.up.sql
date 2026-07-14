BEGIN;

ALTER TABLE training.trainings RENAME COLUMN user_uuid TO user_id;
ALTER TABLE training.trainings ALTER COLUMN user_id TYPE text;

COMMIT;
