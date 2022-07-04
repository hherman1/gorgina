-- name: PutItem :execresult
INSERT INTO catalog
(id, category, brand, color, pattern, title, description, price)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (id) DO UPDATE
SET category=$2, brand=$3, color=$4, pattern=$5, title=$6, description=$7, price=$8;

-- name: ListCatalog :many
SELECT * FROM CATALOG ORDER BY hidden ASC, last_activity DESC NULLS LAST;

-- name: GetCatalog :one
SELECT * FROM CATALOG WHERE id=$1;

-- name: SearchCatalog :many
SELECT * FROM CATALOG WHERE LOWER(title) LIKE '%' || LOWER($1) || '%'
	OR LOWER(description) LIKE '%' || LOWER($1) || '%'
	OR LOWER(color) LIKE '%' || LOWER($1) || '%'
	OR LOWER(category) LIKE '%' || LOWER($1) || '%'
	OR LOWER(brand) LIKE '%' || LOWER($1) || '%'
	OR LOWER(pattern) LIKE '%' || LOWER($1) || '%'
	ORDER BY hidden ASC, last_activity DESC NULLS LAST
	;

-- name: SetHidden :exec
UPDATE catalog SET hidden=$1 WHERE id=$2;

-- name: GetLastUsage :one
SELECT * FROM ACTIVITY WHERE c_id=$1 ORDER BY ts DESC LIMIT 1;

-- name: GetAllUsage :many
SELECT * FROM ACTIVITY WHERE c_id=$1 ORDER BY ts DESC;

-- name: GetUsage :one
SELECT * FROM ACTIVITY WHERE id=$1;

-- name: LogUsage :execresult
INSERT INTO ACTIVITY(id, c_id, ts) values ($1, $2, $3);

-- name: SetUsageNote :execresult
UPDATE activity SET note=$1 WHERE id=$2;

-- name: PutUsage :execresult
UPDATE activity SET note=$1, ts=$2 WHERE id=$3;

-- name: UpdateLastUsed :execresult
UPDATE catalog SET last_activity=$1, last_note=NULL WHERE id=$2;

-- name: UpdateLastNote :execresult
UPDATE catalog SET last_note=$1 WHERE id=$2;

-- name: ListUsage :many
SELECT * FROM ACTIVITY ORDER BY ts DESC;
