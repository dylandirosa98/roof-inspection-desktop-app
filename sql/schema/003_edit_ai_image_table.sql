-- +goose Up
ALTER TABLE ai_images
ADD COLUMN annotations_json TEXT;

-- +goose Down
ALTER TABLE ai_images
DROP COLUMN annotations_json;