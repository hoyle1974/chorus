package dbx

import (
	"context"
	"sync/atomic"

	"github.com/hoyle1974/chorus/db"
	"github.com/jackc/pgx/v5"
)

var conn atomic.Pointer[pgx.Conn]

func getConn() *pgx.Conn {
	if conn.Load() != nil {
		return conn.Load()
	}
	connStr := "host=localhost dbname=chorus user=postgres password=postgres sslmode=disable"

	db, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		panic(err)
	}

	old := conn.Swap(db)
	if old != nil {
		old.Close(context.Background())
	}
	return db
}

func q() *db.Queries {
	return db.New(getConn())
}
