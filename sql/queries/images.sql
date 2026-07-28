-- name: CreateImage :one
INSERT INTO images (width, height, file_size, format, path, data_url, preview_url, project_id )
VALUES (?,?,?,?,?,?,?,?)
RETURNING id, width, height, file_size, format, path, data_url, preview_url, project_id;


-- name: RetrieveImages :many
SELECT width, height, file_size, format, path, data_url, preview_url, id, project_id FROM images
WHERE project_id = ?;

-- name: CreateAiImage :one
INSERT INTO ai_images (image_id, annotations_json)
VALUES (?, ?)
RETURNING id, image_id;

-- name: RetrieveAiImages :many
SELECT
    images.id AS image_id,
    images.path AS image_path,
    images.preview_url AS image_preview_url,
    ai_images.id AS ai_image_id,
    ai_images.annotations_json
FROM images
LEFT JOIN ai_images
    ON images.id = ai_images.image_id
WHERE project_id = ?
ORDER BY ai_images.id IS NOT NULL DESC, images.id;
