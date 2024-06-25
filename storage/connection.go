package storage

import (
	"database/sql"
	"fmt"
	"log"
)

func OpenSql(driverName, url string) (*sql.DB, error) {
	db, err := sql.Open(driverName, url)
	if err != nil {
		fmt.Println(driverName, url)
		log.Println("failed to open database:", err)
		return nil, err
	}

	if err = db.Ping(); err != nil {
		log.Println("Unable to connect to database:", err)
		return nil, err
	}

	return db, err
}