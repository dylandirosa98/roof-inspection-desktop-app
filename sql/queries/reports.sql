-- name: CreateInspectionReport :one
INSERT INTO inspection_reports (project_id, report_number)
VALUES (?, ?)
RETURNING *;

-- name: GetInspectionReport :one
SELECT *
FROM inspection_reports
WHERE id = ?;

-- name: GetInspectionReportByProjectID :one
SELECT *
FROM inspection_reports
WHERE project_id = ?
ORDER BY id DESC
LIMIT 1;

-- name: UpdateInspectionReport :one
UPDATE inspection_reports
SET
    report_number = ?,
    report_title = ?,
    customer_name = ?,
    property_address = ?,
    property_city = ?,
    property_state = ?,
    property_zip = ?,
    inspector_name = ?,
    inspection_date = ?,
    insurance_carrier = ?,
    claim_number = ?,
    date_of_loss = ?,
    summary = ?,
    notes = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: UpdateInspectionReportOutput :exec
UPDATE inspection_reports
SET
    last_generated_pdf_path = ?,
    last_generated_at = ?
WHERE id = ?;

-- name: RetrieveReviewedReportImages :many
SELECT
    images.id AS image_id,
    images.path AS image_path,
    ai_images.edited_annotations_json
FROM images
JOIN ai_images
    ON ai_images.image_id = images.id
WHERE images.project_id = (
    SELECT project_id
    FROM inspection_reports
    WHERE inspection_reports.id = ?
)
AND ai_images.edited_annotations_json IS NOT NULL
AND ai_images.edited_annotations_json <> ''
AND ai_images.review_approved = 1
ORDER BY images.id;
