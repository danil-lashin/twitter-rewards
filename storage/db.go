package storage

import (
	"github.com/danil-lashin/twitter-rewards/config"
	"github.com/syndtr/goleveldb/leveldb"
)

type DB struct {
	ldb *leveldb.DB
}

func NewDB(cfg *config.Config) *DB {
	db, err := leveldb.OpenFile("db/", nil)
	if err != nil {
		panic(err)
	}

	return &DB{
		ldb: db,
	}
}

func (db *DB) IsUserExists(id string) bool {
	_, err := db.ldb.Get([]byte(id), nil)

	return err == nil && err != leveldb.ErrNotFound
}

func (db *DB) AddUser(id string) error {
	return db.ldb.Put([]byte(id), []byte{1}, nil)
}
