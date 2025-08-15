-- +goose up
ALTER TABLE users
ADD COLUMN hashed_password TEXT NOT NULL DEFAULT 'unset';

-- +goose down
ALTER TABLE user
DROP COLUMN hashed_password;