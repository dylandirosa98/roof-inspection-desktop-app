-- name: CreateProject :one
INSERT INTO projects (name, directory)
VALUES (?, ?)
RETURNING id, name, directory;

-- name: RetrieveProjects :many
SELECT id, name, directory,(
        SELECT COUNT(*)
        FROM images
        WHERE images.project_id = projects.id
    ) AS image_count
FROM projects;