package distributed

import (
	"context"
)

// Commands we support:
// Lindex, Linsert, LLen, LPop, LPos, LPush, LPushX, LRange, LRem,
// LSet, LTrim, RPop, RPush, RPushX
type List struct {
	dist Dist
	key  string
}

func (d Dist) BindList(key string, op ...any) List {
	d.commonOps(key, op)
	return List{key: key, dist: d}
}

func (l List) LLindex(index int64) (string, error) {
	return stringCmdWrap(l.dist.conn.LIndex(context.Background(), l.key, index))
}

func (l List) LInsertBefore(value string, elem string) (int64, error) {
	return intCmdWrap(l.dist.conn.LInsertBefore(context.Background(), l.key, value, elem))
}
func (l List) LInsertAfter(value string, elem string) (int64, error) {
	return intCmdWrap(l.dist.conn.LInsertAfter(context.Background(), l.key, value, elem))
}

func (l List) LLen() (int64, error) {
	return intCmdWrap(l.dist.conn.LLen(context.Background(), l.key))
}

func (l List) LPop(values ...interface{}) (string, error) {
	return stringCmdWrap(l.dist.conn.LPop(context.Background(), l.key))
}

func (l List) LPush(values ...interface{}) (int64, error) {
	return intCmdWrap(l.dist.conn.LPush(context.Background(), l.key, values))
}

func (l List) LPushX(values ...interface{}) (int64, error) {
	return intCmdWrap(l.dist.conn.LPushX(context.Background(), l.key, values))
}

func (l List) LRange(start, stop int64) ([]string, error) {
	return stringSliceCmdWrap(l.dist.conn.LRange(context.Background(), l.key, start, stop))
}

func (l List) LRem(count int64, value string) (int64, error) {
	return intCmdWrap(l.dist.conn.LRem(context.Background(), l.key, count, value))
}

func (l List) LSet(index int64, value interface{}) (string, error) {
	return statusCmdWrap(l.dist.conn.LSet(context.Background(), l.key, index, value))
}

func (l List) LTrim(start, stop int64) (string, error) {
	return statusCmdWrap(l.dist.conn.LTrim(context.Background(), l.key, start, stop))
}

func (l List) RPop(values ...interface{}) (string, error) {
	return stringCmdWrap(l.dist.conn.RPop(context.Background(), l.key))
}

func (l List) RPush(values ...interface{}) (int64, error) {
	return intCmdWrap(l.dist.conn.RPush(context.Background(), l.key, values))
}

func (l List) RPushX(values ...interface{}) (int64, error) {
	return intCmdWrap(l.dist.conn.RPushX(context.Background(), l.key, values))
}
