-- Drop indexes
DROP INDEX idx_unique_room_key_valiue;
DROP INDEX idx_room_membership_room_uuid;
DROP INDEX idx_room_membership_connection_uuid;
DROP INDEX idx_unique_true_value_constraint;
DROP INDEX idx_unique_true_value;
DROP INDEX idx_machine_primary;

-- Drop functions
DROP FUNCTION notify_machine_update_trigger();

-- Drop triggers
DROP TRIGGER machines_trigger ON machines;

-- Drop tables
DROP TABLE room_membership;
DROP TABLE room_data;
DROP TABLE connections;
DROP TABLE rooms;
DROP TABLE machines;
