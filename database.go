package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

type Database interface {
	Close()
}

type database struct {
	db *sql.DB
}

func MakeDatabase() Database {
	d := &database{}

	db, err := sql.Open("mysql", "server:mysql@/mmgorogue")
	if err != nil {
		log.Fatal(err)
	}

	d.db = db
	return d
}

func (d *database) Close() {
	d.db.Close()
}
