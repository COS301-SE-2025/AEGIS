package db

import "database/sql"

// Define an interface
type Database interface {
    Query(query string, args ...interface{}) (*sql.Rows, error)
    Exec(query string, args ...interface{}) (sql.Result, error)
    Ping() error
}

// Use the interface instead of *sql.DB
type PostgresDB struct {
    DB *sql.DB
}

func (p *PostgresDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
    return p.DB.Query(query, args...)
}

func (p *PostgresDB) Exec(query string, args ...interface{}) (sql.Result, error) {
    return p.DB.Exec(query, args...)
}

func (p *PostgresDB) Ping() error {
    return p.DB.Ping()
}