-- name: CreateImage :one
INSERT INTO images (width, height, file_size, format, path, data_url, preview_url, project_id )
VALUES (?,?,?,?,?,?,?,?)
RETURNING id, width, height, file_size, format, path, data_url, preview_url, project_id;