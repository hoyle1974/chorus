
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

-- name: DeleteRoom :exec
DELETE FROM rooms
WHERE uuid = $1;





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