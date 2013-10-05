package main

import (
	"errors"
	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/thrsafe"
	"log"
)

type Database interface {
	Close()
	Authenticate(name string, password string) bool
	CreateUser(name string, password string, email string) (bool, error)
	UserExists(name string) bool
}

type database struct {
	db		mysql.Conn
	authStmt	mysql.Stmt
	createUserStmt	mysql.Stmt
	userExistsStmt	mysql.Stmt
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func (d *database) prepareStatements() {
	var err error = nil
	db := d.db

	d.authStmt, err = db.Prepare("CALL authenticate(?, ?)")
	checkError(err)
	d.createUserStmt, err = db.Prepare("CALL create_user(?, ?, ?)")
	checkError(err)
	d.userExistsStmt, err = db.Prepare("CALL user_exists(?)")
	checkError(err)
}

func (d *database) terminateStatements() {
	d.authStmt.Delete()
	d.createUserStmt.Delete()
	d.userExistsStmt.Delete()
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

func eatRemainingResults(res mysql.Result) {
	var err error = nil
	for !res.StatusOnly() {
		res, err = res.NextResult()
		checkError(err)
		if res == nil {
			log.Fatal("nil query result!")
		}
	}

}

func (d *database) Authenticate(name string, password string) bool {
	row, res, err := d.authStmt.ExecFirst(name, password)
	checkError(err)
	
	success, err := row.BoolErr(0)
	checkError(err)
	
	eatRemainingResults(res)
	return success
}


func (d *database) CreateUser(name string, password string, email string) (bool, error) {
	row, res, err := d.createUserStmt.ExecFirst(name, password, email)
	checkError(err)

	msg := row.Str(0)
	eatRemainingResults(res)
	
	err = nil
	if msg != "OK" {
		err = errors.New(msg)
	}

	return msg == "OK", err
}

func (d *database) UserExists(name string) bool {
	row, res, err := d.userExistsStmt.ExecFirst(name)
	checkError(err)

	exists, err := row.BoolErr(0)
	checkError(err)

	eatRemainingResults(res)
	return exists
}
