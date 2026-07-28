-- +goose Up
ALTER TABLE ai_images DROP COLUMN width;
ALTER TABLE ai_images DROP COLUMN height;
ALTER TABLE ai_images DROP COLUMN file_size;
ALTER TABLE ai_images DROP COLUMN format;
ALTER TABLE ai_images DROP COLUMN path;
ALTER TABLE ai_images DROP COLUMN data_url;
ALTER TABLE ai_images DROP COLUMN preview_url;

-- +goose Down
ALTER TABLE ai_images ADD COLUMN width INTEGER;
ALTER TABLE ai_images ADD COLUMN height INTEGER;
ALTER TABLE ai_images ADD COLUMN file_size INTEGER;
ALTER TABLE ai_images ADD COLUMN format TEXT;
ALTER TABLE ai_images ADD COLUMN path TEXT;
ALTER TABLE ai_images ADD COLUMN data_url TEXT;
ALTER TABLE ai_images ADD COLUMN preview_url TEXT;
