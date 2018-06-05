package broker

import (
	"context"
	"database/sql"
	"time"

	"github.com/golang/glog"

	_ "gopkg.in/rana/ora.v4"
)

//Creates user in the DB using the provided information
func createUser(username, password, sysconn string) error {
	db, err := sql.Open("ora", sysconn)
	if err != nil {
		glog.Errorf("failed to get instance ID from cloud provider: %v", err)
	}
	if err = db.Ping(); err != nil {
		glog.Errorf("failed to ping the DB: %v", err)
	}
	defer db.Close()

	// Set timeout (Go 1.8)
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	// Set prefetch count (Go 1.8)
	// ctx = ora.WithStmtCfg(ctx, ora.Cfg().StmtCfg.SetPrefetchCount(50000))
	//Create User
	result, err := db.ExecContext(ctx, "CREATE USER "+username+" IDENTIFIED BY "+password)
	if err != nil {
		glog.Errorf("failed to create user %v: %v", username, err)
	}
	glog.V(4).Infof("Result: %s", result)

	//Grant permissions to USER
	result, err = db.ExecContext(ctx, "GRANT CONNECT, CREATE SESSION, RESOURCE TO "+username)
	if err != nil {
		glog.Errorf("failed to grant privileges to %v: %v", username, err)
		return err
	}
	glog.V(4).Infof("Result: %s", result)

	return nil
}

//Create Table
func createTable(connURI, tablename, tableschema string) error {

	db, err := sql.Open("ora", connURI)
	if err != nil {
		glog.Errorf("failed to get instance ID from cloud provider: %v", err)
	}
	if err = db.Ping(); err != nil {
		glog.Errorf("failed to ping the DB: %v", err)
	}
	defer db.Close()

	// Set timeout (Go 1.8)
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	// Set prefetch count (Go 1.8)
	// ctx = ora.WithStmtCfg(ctx, ora.Cfg().StmtCfg.SetPrefetchCount(50000))

	//Create Table
	//tablename - Persons
	//tableschema - PersonID int, LastName varchar(255), FirstName varchar(255), Address varchar(255), City varchar(255)
	//SQL -  CREATE TABLE tablename ( tableschema )
	result, err := db.ExecContext(ctx, "CREATE TABLE "+tablename+" "+"( "+tableschema+" )")
	if err != nil {
		glog.Errorf("failed to create table %v: %v", tablename, err)
	}
	glog.V(4).Infof("Result: %s", result)

	return nil
}

//Delete User in the DB
func deleteUser(username, sysconn string) error {
	db, err := sql.Open("ora", sysconn)
	if err != nil {
		glog.Errorf("failed to get instance ID from cloud provider: %v", err)
	}
	if err = db.Ping(); err != nil {
		glog.Errorf("failed to ping the DB: %v", err)
	}
	defer db.Close()

	// Set timeout (Go 1.8)
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	// Set prefetch count (Go 1.8)
	// ctx = ora.WithStmtCfg(ctx, ora.Cfg().StmtCfg.SetPrefetchCount(50000))
	result, err := db.ExecContext(ctx, "DROP USER "+"\""+username+"\""+"cascade")
	if err != nil {
		glog.Errorf("failed to delete user %v: %v", username, err)
	}
	glog.V(4).Infof("Result: %s", result)

	return nil
}
