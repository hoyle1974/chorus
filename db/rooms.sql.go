// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: rooms.sql

package db

import (
	"context"
)

const addRoomMember = `-- name: AddRoomMember :exec
INSERT INTO room_membership (
    connection_uuid, room_uuid
) VALUES (
    $1, $2
)
`

type AddRoomMemberParams struct {
	ConnectionUuid string
	RoomUuid       string
}

// CREATE TABLE room_membership (
//
//	connection_uuid TEXT NOT NULL REFERENCES connections(uuid) ON DELETE CASCADE,
//	room_uuid TEXT NOT NULL REFERENCES rooms(uuid) ON DELETE CASCADE
//
// );
func (q *Queries) AddRoomMember(ctx context.Context, arg AddRoomMemberParams) error {
	_, err := q.db.Exec(ctx, addRoomMember, arg.ConnectionUuid, arg.RoomUuid)
	return err
}

const createRoom = `-- name: CreateRoom :exec
INSERT INTO rooms (
    uuid, machine_uuid, name, script, destroy_on_orphan
) VALUES (
    $1, $2, $3, $4, $5
)
`

type CreateRoomParams struct {
	Uuid            string
	MachineUuid     string
	Name            string
	Script          string
	DestroyOnOrphan bool
}

func (q *Queries) CreateRoom(ctx context.Context, arg CreateRoomParams) error {
	_, err := q.db.Exec(ctx, createRoom,
		arg.Uuid,
		arg.MachineUuid,
		arg.Name,
		arg.Script,
		arg.DestroyOnOrphan,
	)
	return err
}

const deleteRoom = `-- name: DeleteRoom :exec
DELETE FROM rooms
WHERE uuid = $1
`

func (q *Queries) DeleteRoom(ctx context.Context, uuid string) error {
	_, err := q.db.Exec(ctx, deleteRoom, uuid)
	return err
}

const getMembershipByConnection = `-- name: GetMembershipByConnection :many
SELECt room_uuid from room_membership where connection_uuid = $1
`

func (q *Queries) GetMembershipByConnection(ctx context.Context, connectionUuid string) ([]string, error) {
	rows, err := q.db.Query(ctx, getMembershipByConnection, connectionUuid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []string
	for rows.Next() {
		var room_uuid string
		if err := rows.Scan(&room_uuid); err != nil {
			return nil, err
		}
		items = append(items, room_uuid)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getOrphanedRooms = `-- name: GetOrphanedRooms :many
SELECT uuid, machine_uuid, name, script, destroy_on_orphan, created_at, last_updated FROM rooms
WHERE machine_uuid NOT IN (
SELECT uuid
FROM machines
WHERE last_updated < NOW() - INTERVAL '5 seconds'
)
`

func (q *Queries) GetOrphanedRooms(ctx context.Context) ([]Room, error) {
	rows, err := q.db.Query(ctx, getOrphanedRooms)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Room
	for rows.Next() {
		var i Room
		if err := rows.Scan(
			&i.Uuid,
			&i.MachineUuid,
			&i.Name,
			&i.Script,
			&i.DestroyOnOrphan,
			&i.CreatedAt,
			&i.LastUpdated,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getRoomMembers = `-- name: GetRoomMembers :many
SELECT connection_uuid FROM room_membership where room_uuid = $1
`

func (q *Queries) GetRoomMembers(ctx context.Context, roomUuid string) ([]string, error) {
	rows, err := q.db.Query(ctx, getRoomMembers, roomUuid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []string
	for rows.Next() {
		var connection_uuid string
		if err := rows.Scan(&connection_uuid); err != nil {
			return nil, err
		}
		items = append(items, connection_uuid)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getRooms = `-- name: GetRooms :many

SELECT uuid, machine_uuid, name, script, destroy_on_orphan, created_at, last_updated FROM rooms
`

// CREATE TABLE rooms (
//
//	uuid TEXT PRIMARY KEY,
//	machine_uuid TEXT NOT NULL REFERENCES machines(uuid) ,
//	name TEXT NOT NULL,
//	script TEXT NOT NULL,
//	destroy_on_orphan BOOLEAN NOT NULL
//
// );
func (q *Queries) GetRooms(ctx context.Context) ([]Room, error) {
	rows, err := q.db.Query(ctx, getRooms)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Room
	for rows.Next() {
		var i Room
		if err := rows.Scan(
			&i.Uuid,
			&i.MachineUuid,
			&i.Name,
			&i.Script,
			&i.DestroyOnOrphan,
			&i.CreatedAt,
			&i.LastUpdated,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getRoomsByMachine = `-- name: GetRoomsByMachine :many
SELECT uuid, machine_uuid, name, script, destroy_on_orphan, created_at, last_updated FROM rooms WHERE machine_uuid = $1
`

func (q *Queries) GetRoomsByMachine(ctx context.Context, machineUuid string) ([]Room, error) {
	rows, err := q.db.Query(ctx, getRoomsByMachine, machineUuid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Room
	for rows.Next() {
		var i Room
		if err := rows.Scan(
			&i.Uuid,
			&i.MachineUuid,
			&i.Name,
			&i.Script,
			&i.DestroyOnOrphan,
			&i.CreatedAt,
			&i.LastUpdated,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const removeRoomMember = `-- name: RemoveRoomMember :exec
DELETE FROM room_membership 
WHERE connection_uuid = $1 AND room_uuid = $2
`

type RemoveRoomMemberParams struct {
	ConnectionUuid string
	RoomUuid       string
}

func (q *Queries) RemoveRoomMember(ctx context.Context, arg RemoveRoomMemberParams) error {
	_, err := q.db.Exec(ctx, removeRoomMember, arg.ConnectionUuid, arg.RoomUuid)
	return err
}

const setRoomOwner = `-- name: SetRoomOwner :exec
UPDATE rooms 
SET
    machine_uuid = $2
WHERE 
    uuid = $1
AND
    machine_uuid = $3
`

type SetRoomOwnerParams struct {
	Uuid          string
	MachineUuid   string
	MachineUuid_2 string
}

func (q *Queries) SetRoomOwner(ctx context.Context, arg SetRoomOwnerParams) error {
	_, err := q.db.Exec(ctx, setRoomOwner, arg.Uuid, arg.MachineUuid, arg.MachineUuid_2)
	return err
}
