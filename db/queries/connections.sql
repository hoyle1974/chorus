-- name: GetConnections :many
SELECT * FROM connections;


--CREATE TABLE connections (
--    uuid TEXT PRIMARY KEY,
--    machine_uuid TEXT NOT NULL REFERENCES machines(uuid) ,
--    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
--    last_updated TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
--);

-- name: CreateConnection :exec
INSERT INTO connections (
    uuid, machine_uuid
) VALUES (
    $1, $2
);

-- name: DeleteConnection :exec
DELETE FROM connections 
WHERE uuid = $1;

-- name: FindMachine :one
SELECT * FROM connections
WHERE uuid = $1;

-- name: TouchConnection :exec
UPDATE connections 
SET last_updated = now()
WHERE uuid = $1;