-- +goose Up
ALTER TABLE ai_images ADD COLUMN edited_annotations_json TEXT;

-- +goose Down
ALTER TABLE ai_images DROP COLUMN edited_annotations_json;
