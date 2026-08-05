-- +goose Up
ALTER TABLE ai_images
ADD COLUMN review_approved INTEGER NOT NULL DEFAULT 0
CHECK (review_approved IN (0, 1));

ALTER TABLE ai_images
ADD COLUMN reviewed_at TEXT;

-- +goose Down
ALTER TABLE ai_images
DROP COLUMN reviewed_at;

ALTER TABLE ai_images
DROP COLUMN review_approved;
