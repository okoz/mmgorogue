package main

import (
	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/thrsafe"
	"log"
)

type Database interface {
	Close()
	Authenticate(userName string, password string) bool
}

type database struct {
	db		mysql.Conn
	authStatement	mysql.Stmt
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func (d *database) prepareStatements() {
	var err error = nil
	db := d.db

	d.authStatement, err = db.Prepare("CALL authenticate(?, ?)")
	checkError(err)
}

func (d *database) terminateStatements() {
	d.authStatement.Delete()
}

func MakeDatabase() Database {
	d := &database{}

	db := mysql.New("tcp", "", "127.0.0.1:3306", "server", "mysql", "mmgorogue")
	err := db.Connect()
	checkError(err)

	d.db = db
	d.prepareStatements()
	return d
}

func (d *database) Close() {
	d.terminateStatements()
	d.db.Close()
}

func (d *database) Authenticate(name string, password string) bool {
	row, res, err := d.authStatement.ExecFirst(name, password)
	checkError(err)

	success, err := row.BoolErr(0)
	checkError(err)

	for !res.StatusOnly() {
		res, err = res.NextResult()
		checkError(err)
		if res == nil {
			log.Fatal("nil query result!")
		}
	}

	return success
}

