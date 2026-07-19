-- +goose Up
CREATE TABLE ai_images(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    width INTEGER,
    height INTEGER,
    file_size INTEGER,
    format TEXT,
    path TEXT NOT NULL,
    data_url TEXT,
    preview_url TEXT,
    image_id INTEGER NOT NULL UNIQUE,
    FOREIGN KEY (image_id) REFERENCES images(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE ai_images;