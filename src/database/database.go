package database

import (
	"database/sql"
)

func New() (*sql.DB, error) {
	// db, err := sql.Open("sqlite3", "database.db")
	// if err != nil {
	// 	return nil, err
	// }

	// defer db.Close()
	return nil, nil
}

func prepareTables(db *sql.DB) {
	// errors := make(chan error, 1)

	// db.Exec("CREATE TABLE IF NOT EXIST Images ();")
}
