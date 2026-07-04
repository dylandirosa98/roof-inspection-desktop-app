-- +goose Up
CREATE TABLE projects(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    directory TEXT NOT NULL
);

CREATE TABLE images(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    width INTEGER,
    height INTEGER,
    file_size INTEGER,
    format TEXT,
    path TEXT NOT NULL,
    data_url TEXT,
    preview_url TEXT,
    project_id INTEGER NOT NULL,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE images;
DROP TABLE projects;
