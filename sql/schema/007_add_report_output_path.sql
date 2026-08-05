-- +goose Up
ALTER TABLE inspection_reports
ADD COLUMN last_generated_pdf_path TEXT NOT NULL DEFAULT '';

ALTER TABLE inspection_reports
ADD COLUMN last_generated_at TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE inspection_reports
DROP COLUMN last_generated_at;

ALTER TABLE inspection_reports
DROP COLUMN last_generated_pdf_path;
