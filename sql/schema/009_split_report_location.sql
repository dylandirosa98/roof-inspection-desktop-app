-- +goose Up
ALTER TABLE inspection_reports
ADD COLUMN property_city TEXT NOT NULL DEFAULT '';

ALTER TABLE inspection_reports
ADD COLUMN property_state TEXT NOT NULL DEFAULT '';

ALTER TABLE inspection_reports
ADD COLUMN property_zip TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE inspection_reports
DROP COLUMN property_zip;

ALTER TABLE inspection_reports
DROP COLUMN property_state;

ALTER TABLE inspection_reports
DROP COLUMN property_city;
