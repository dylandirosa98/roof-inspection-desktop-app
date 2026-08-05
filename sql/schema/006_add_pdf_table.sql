-- +goose Up
CREATE TABLE inspection_reports (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,

    report_number TEXT NOT NULL UNIQUE,
    report_title TEXT NOT NULL DEFAULT 'Roof Inspection and Photo Documentation Report',

    customer_name TEXT NOT NULL DEFAULT '',
    property_address TEXT NOT NULL DEFAULT '',
    city_state_zip TEXT NOT NULL DEFAULT '',

    inspector_name TEXT NOT NULL DEFAULT '',
    inspection_date TEXT NOT NULL DEFAULT '',

    insurance_carrier TEXT NOT NULL DEFAULT '',
    claim_number TEXT NOT NULL DEFAULT '',
    date_of_loss TEXT NOT NULL DEFAULT '',

    summary TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',

    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE inspection_reports;
