
--CREATE TABLE rooms (
--    uuid TEXT PRIMARY KEY,
--    machine_uuid TEXT NOT NULL REFERENCES machines(uuid) ,
--    name TEXT NOT NULL,
--    script TEXT NOT NULL,
--    destroy_on_orphan BOOLEAN NOT NULL
--);

-- name: GetRooms :many
SELECT * FROM rooms;

-- name: CreateRoom :exec
INSERT INTO rooms (
    uuid, machine_uuid, name, script, destroy_on_orphan
) VALUES (
    $1, $2, $3, $4, $5
);

-- name: SetRoomOwner :exec
UPDATE rooms 
SET
    machine_uuid = $2
WHERE 
    uuid = $1
AND
    machine_uuid = $3;

-- name: DeleteRoom :exec
DELETE FROM rooms
WHERE uuid = $1;


-- name: GetOrphanedRooms :many
SELECT * FROM rooms
WHERE machine_uuid NOT IN (
SELECT uuid
FROM machines
WHERE last_updated < NOW() - INTERVAL '5 seconds'
);

-- name: GetRoomsByMachine :many
SELECT * FROM rooms WHERE machine_uuid = $1;


--CREATE TABLE room_membership (
--    connection_uuid TEXT NOT NULL REFERENCES connections(uuid) ON DELETE CASCADE,
--    room_uuid TEXT NOT NULL REFERENCES rooms(uuid) ON DELETE CASCADE
--);
-- name: AddRoomMember :exec
INSERT INTO room_membership (
    connection_uuid, room_uuid
) VALUES (
    $1, $2
);

-- name: RemoveRoomMember :exec
DELETE FROM room_membership 
WHERE connection_uuid = $1 AND room_uuid = $2;

-- name: GetRoomMembers :many
SELECT connection_uuid FROM room_membership where room_uuid = $1;

-- name: GetMembershipByConnection :many
SELECt room_uuid from room_membership where connection_uuid = $1;