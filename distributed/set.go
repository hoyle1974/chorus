package distributed

import "context"

// Commands we support: SAdd, SCard, SIsMember, SMembers, SMIsMember, SPop
// SRandMember, SRem, SScan,
type Set struct {
	dist Dist
	key  string
}

func (d Dist) BindSet(key string, op ...any) Set {
	d.commonOps(key, op)
	return Set{key: key, dist: d}
}

func (l Set) SAdd(members ...interface{}) (int64, error) {
	return intCmdWrap(l.dist.conn.SAdd(context.Background(), l.key, members))
}

func (l Set) SCard() (int64, error) {
	return intCmdWrap(l.dist.conn.SCard(context.Background(), l.key))
}

func (l Set) SIsMember(member string) (bool, error) {
	return boolCmdWrap(l.dist.conn.SIsMember(context.Background(), l.key, member))
}

func (l Set) SMembers() ([]string, error) {
	return stringSliceCmdWrap(l.dist.conn.SMembers(context.Background(), l.key))
}

func (l Set) SMIsMember(members []string) ([]bool, error) {
	return boolSliceCmdWrap(l.dist.conn.SMIsMember(context.Background(), l.key, members))
}

func (l Set) SPop() (string, error) {
	return stringCmdWrap(l.dist.conn.SPop(context.Background(), l.key))
}

func (l Set) SRandMember() (string, error) {
	return stringCmdWrap(l.dist.conn.SRandMember(context.Background(), l.key))
}

func (l Set) SRem(members ...interface{}) (int64, error) {
	return intCmdWrap(l.dist.conn.SRem(context.Background(), l.key, members))
}

func (l Set) SScan(cursor uint64, match string, count int64) ([]string, uint64, error) {
	return scanCmdWrap(l.dist.conn.SScan(context.Background(), l.key, cursor, match, count))
}
