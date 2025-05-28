package db

import (
    "errors"
    "database/sql"
)

// Mock DB struct
type MockDB struct{}

func (m *MockDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
    return nil, errors.New("mock query: no rows")
}

func (m *MockDB) Exec(query string, args ...interface{}) (sql.Result, error) {
    return nil, errors.New("mock exec: operation not permitted")
}

func (m *MockDB) Ping() error {
    return nil // Simulate successful connection
}