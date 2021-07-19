package main

import (
	"database/sql"
	"log"
	"unicode"
)

type Database struct {
	Name           string
	db             *sql.DB
	PartitionStart rune
	PartitionEnd   rune
}

func NewDatabase(name string, connectionString string, partitionStart rune, partitionEnd rune) (*Database, error) {
	db, err := sql.Open("sqlite3", connectionString)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	return &Database{
		Name:           name,
		db:             db,
		PartitionStart: unicode.ToUpper(partitionStart),
		PartitionEnd:   unicode.ToUpper(partitionEnd),
	}, nil
}

func (i *Database) GetConnection() *sql.DB {
	return i.db
}

func (i *Database) Close() {
	i.db.Close()
}
