package utils

import (
	"context"
	"database/sql"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type PGXPool interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
	Begin(ctx context.Context) (pgx.Tx, error)
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
	Close()
}

func ConnectDBPool(connString string) *pgxpool.Pool {
	dbpool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		panic(err)
	}

	return dbpool
}

func ConnectDB(connString string) *sql.DB {
	dbc, err := sql.Open("pgx", connString)
	if err != nil {
		panic(err)
	}
	// Maximum Idle Connections
	dbc.SetMaxIdleConns(10)
	// Maximum Open Connections
	dbc.SetMaxOpenConns(20)
	// Idle Connection Timeout
	dbc.SetConnMaxIdleTime(15 * time.Second)
	// Connection Lifetime
	dbc.SetConnMaxLifetime(60 * time.Second)

	err = dbc.Ping()
	if err != nil {
		panic(err)
	}

	return dbc
}
