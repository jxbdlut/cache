package cache

import (
	"github.com/jinzhu/gorm"
)

type DataBase struct {
	connName    string
	databases   map[string]*gorm.DB
	readDbName  string
	writeDbName string
}

func (d *DataBase) SetReadDbName(name string) {
	d.readDbName = name
}

func (d *DataBase) SetWriteName(name string) {
	d.writeDbName = name
}

func (d *DataBase) SetDatabase(name string, db *gorm.DB) {
	if d.databases == nil {
		d.databases = make(map[string]*gorm.DB)
	}
	d.databases[name] = db
	if name == "default" {
		d.readDbName = name
		d.writeDbName = name
	}
}

func (d *DataBase) Db(name string) *DataBase {
	d.connName = name
	return d
}

func (d *DataBase) GetReadDb() *gorm.DB {
	dbConnectName := d.readDbName
	if d.connName != "" {
		dbConnectName = d.connName
	}
	if db, ok := d.databases[dbConnectName]; ok {
		return db
	}
	log.Panicf("database name %v not exists! %v", dbConnectName)
	return nil
}

func (d *DataBase) GetWriteDb() *gorm.DB {
	dbConnectName := d.writeDbName
	if d.connName != "" {
		dbConnectName = d.connName
	}
	if db, ok := d.databases[dbConnectName]; ok {
		return db
	}
	log.Panicf("database name %v not exists!", dbConnectName)
	return nil
}