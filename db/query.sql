-- name: PutItem :execresult
INSERT INTO catalog
(id, img, title, description)
VALUES ($1, $2, $3, $4)
ON CONFLICT (id) DO UPDATE
SET img = $2, title = $3, description = $4;