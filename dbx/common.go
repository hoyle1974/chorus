package dbx

import (
	"context"
	"sync/atomic"

	"github.com/hoyle1974/chorus/db"
	"github.com/jackc/pgx/v5"
)

var conn atomic.Pointer[pgx.Conn]

func NewConn() (*pgx.Conn, error) {
	connStr := "host=localhost user=postgres password=postgres sslmode=disable"

	return pgx.Connect(context.Background(), connStr)
}

func GetConn() *pgx.Conn {
	if conn.Load() != nil {
		return conn.Load()
	}

	db, err := NewConn()
	if err != nil {
		panic(err)
	}

	old := conn.Swap(db)
	if old != nil {
		old.Close(context.Background())
	}
	return db
}

type DBX struct {
}

func Dbx() DBX {
	return DBX{}
}

type QueriesX struct {
	q *db.Queries
}

func (dbx DBX) Queries(q *db.Queries) QueriesX {
	return QueriesX{q: q}
}
