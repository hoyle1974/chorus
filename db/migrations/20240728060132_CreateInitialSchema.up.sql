-- Each machine listens to this channel
-- When a machine is marked as the monitor, all other machines
-- watch for last_updated changes on that row.  If that row doesn't have an update for more than 5 seconds we assume the machine is
-- no longer the monitor and another machine will try to take it's place
-- The machine that is the monitor will 
-- 	- watch the machines table and clean it up if machines go away
-- 	- when machines go, the rooms & connections table is watched and cleaned up as needed
CREATE OR REPLACE FUNCTION notify_machine_update_trigger() RETURNS trigger AS $$
BEGIN
  PERFORM pg_notify('machines', 'monitor data updated');
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- When a server starts it adds a row to this table
-- some process monitors this table and cleans it up if needed
CREATE TABLE machines (
    uuid TEXT PRIMARY KEY,
    monitor BOOLEAN NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_machine_primary ON machines (monitor);

-- Only allow one row to be true
CREATE INDEX idx_unique_true_value ON machines (monitor) WHERE monitor = true;
CREATE UNIQUE INDEX idx_unique_true_value_constraint ON machines (monitor) WHERE monitor = true;

-- CREATE TRIGGER machines_trigger_monitor
-- AFTER UPDATE OF last_updated ON machines
-- FOR (monitor = true) ROW
-- EXECUTE FUNCTION notify_machine_update_trigger();

CREATE TRIGGER machines_trigger_monitor
AFTER UPDATE ON machines
FOR EACH ROW
WHEN (OLD.last_updated IS DISTINCT FROM NEW.last_updated)
EXECUTE FUNCTION notify_machine_update_trigger();

-- This stores rooms
-- When the machine table is cleaned it either deletes rooms or rehomes them
CREATE TABLE rooms (
    uuid TEXT PRIMARY KEY,
    machine_uuid TEXT NOT NULL REFERENCES machines(uuid),
    name TEXT NOT NULL,
    script TEXT NOT NULL,
    destroy_on_orphan BOOLEAN NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE room_data (
    room_uuid TEXT REFERENCES rooms(uuid),
    key TEXT NOT NULL,
    value TEXT
);
CREATE UNIQUE INDEX idx_unique_room_key_valiue ON room_data (room_uuid,key);

-- This stores connections
CREATE TABLE connections (
    uuid TEXT PRIMARY KEY,
    machine_uuid TEXT NOT NULL REFERENCES machines(uuid) ,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE room_membership (
    connection_uuid TEXT NOT NULL REFERENCES connections(uuid) ON DELETE CASCADE,
    room_uuid TEXT NOT NULL REFERENCES rooms(uuid) ON DELETE CASCADE
);

CREATE INDEX idx_room_membership_connection_uuid ON room_membership (connection_uuid);
CREATE INDEX idx_room_membership_room_uuid ON room_membership (room_uuid);
