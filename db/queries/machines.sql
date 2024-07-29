-- name: GetMachines :many
SELECT * FROM machines;

-- name: CreateMachine :exec
INSERT INTO machines (
    uuid, monitor
) VALUES (
    $1, false
);

-- name: DeleteMachine :exec
DELETE FROM machines
WHERE uuid = $1;

-- name: SetMachineAsMonitor :exec
UPDATE machines 
SET monitor=true
WHERE uuid = $1;

-- name: UpdateMachine :exec
UPDATE machines
SET last_updated = NOW()
WHERE uuid = $1;


-- name: GetMonitor :one
SELECT uuid FROM machines
WHERE monitor = true;

-- name: GetExpiredMachines :many
SELECT uuid FROM machines
WHERE last_updated < NOW() - INTERVAL $1 second;


