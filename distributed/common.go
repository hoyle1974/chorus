package distributed

import (
	"github.com/redis/go-redis/v9"
)

func stringSliceCmdWrap(r *redis.StringSliceCmd) ([]string, error) {
	return r.Val(), r.Err()
}
func stringCmdWrap(r *redis.StringCmd) (string, error) {
	return r.Val(), r.Err()
}
func intCmdWrap(r *redis.IntCmd) (int64, error) {
	return r.Val(), r.Err()
}
func statusCmdWrap(r *redis.StatusCmd) (string, error) {
	return r.Val(), r.Err()
}
func boolCmdWrap(r *redis.BoolCmd) (bool, error) {
	return r.Val(), r.Err()
}
func boolSliceCmdWrap(r *redis.BoolSliceCmd) ([]bool, error) {
	return r.Val(), r.Err()
}
func scanCmdWrap(r *redis.ScanCmd) ([]string, uint64, error) {
	keys, cursorId := r.Val()
	return keys, cursorId, r.Err()
}
