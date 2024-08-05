-- name: GetMachines :many
SELECT * FROM machines;

-- name: GetMachinesByType :many
SELECt * FROM machines where machine_type = $1;



-- name: CreateMachine :exec
INSERT INTO machines (
    uuid, machine_type
) VALUES (
    $1, $2
);

-- name: DeleteMachine :exec
DELETE FROM machines
WHERE uuid = $1;

-- name: UpdateMachine :exec
UPDATE machines
SET last_updated = NOW()
WHERE uuid = $1;


-- name: TouchMachine :exec
UPDATE machines 
SET last_updated = now()
WHERE uuid = $1;

-- name: GetMachine :one
SELECT * FROM machines 
WHERE uuid=$1;

-- name: CreateLeader :exec
INSERT INTO machine_type_leader (
    machine_uuid
) VALUES (
    $1
);

-- name: DeleteLeader :exec
DELETE FROM machine_type_leader
WHERE machine_uuid = $1;

-- name: GetLeaderForType :one
SELECT machines.uuid                                                                                                                                                                            
FROM machine_type_leader, machines                                                                                                                                                                            
WHERE machines.machine_type = $1                                                                                                                                                                         
AND machine_type_leader.machine_uuid = machines.uuid;

-- name: SetMachineAsLeader :exec
INSERT INTO machine_type_leader (
    machine_uuid
) VALUES (
    $1
);

-- name: GetMachineLeaderCountByType :one
SELECT COUNT(*) 
FROM machine_type_leader, machines                                                                                                                                                                            
WHERE machines.machine_type = $1                                                                                                                                                                         
AND machine_type_leader.machine_uuid = machines.uuid;

