-- name: CreateProject :one
INSERT INTO projects (name, directory)
VALUES (?, ?)
RETURNING id, name, directory;