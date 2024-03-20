package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rus-sharafiev/go-rest/common/exception"
)

type Postgres struct {
	pool *pgxpool.Pool
}

// -- Create instance -------------------------------------------------------------

var (
	connectOnce sync.Once
	Instance    Postgres
)

func Connect(dbName string) {
	connectOnce.Do(func() {

		pool, err := pgxpool.New(context.Background(), "postgres:///"+dbName)
		if err != nil {
			log.Fatalf("\x1b[2mPostgreSQL:\x1b[0m\x1b[31m unable to create database connection: %s \x1b[0m\n\n", err.Error())
		}

		fmt.Println("\x1b[32mСonnection to the database has been established\x1b[0m")
		Instance = Postgres{pool}
	})
}

// -- Methods ---------------------------------------------------------------------

func (p *Postgres) Query(query *string, args ...any) (pgx.Rows, error) {
	rows, err := p.pool.Query(context.Background(), *query, args...)
	return rows, err
}

func (p *Postgres) QueryRow(query *string, args ...any) pgx.Row {
	return p.pool.QueryRow(context.Background(), *query, args...)
}

func (p *Postgres) PgxPoolClose() {
	p.pool.Close()
}

// Returns JSON string serialized by postgres
func (p *Postgres) JsonString(w http.ResponseWriter, query *string, args ...any) (string, error) {
	var result sql.NullString
	if err := p.pool.QueryRow(context.Background(), *query, args...).Scan(&result); err != nil {
		if err == pgx.ErrNoRows {
			return "null", nil
		} else {
			return "", err
		}
	}
	return result.String, nil
}

// Writes JSON string serialized by postgres to provided http.ResponseWriter
func (p *Postgres) WriteJsonString(w http.ResponseWriter, query *string, args ...any) {
	var result sql.NullString
	if err := p.pool.QueryRow(context.Background(), *query, args...).Scan(&result); err != nil {
		exception.PgxNoRows(w, err)
		return
	}
	fmt.Fprint(w, result.String)
}
