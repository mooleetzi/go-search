package storage

import (
	"fmt"
	"go-search/log"
	"go-search/util"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

var (
	idGenerator = &util.AtomicCounter{}
)

type Storage struct {
	db   *leveldb.DB
	path string
	id   int32
}

// Create database
func Create(path string) (*Storage, error) {

	option := &opt.Options{
		Filter: filter.NewBloomFilter(12),
	}

	db, err := leveldb.OpenFile(path, option)
	if err != nil {
		log.Infof("Create database in [%s] thorw err[%v]", path, err)
		return nil, err
	}
	idx := idGenerator.Load()
	idGenerator.Inc()
	return &Storage{
		db:   db,
		path: path,
		id:   idx,
	}, nil
}

func (s *Storage) Get(key []byte) ([]byte, error) {

	value, err := s.db.Get(key, nil)
	if err != nil {
		log.Infof("%s get key[%v] thorw err[%v]", s, key, err)
	}
	return value, err
}

func (s *Storage) Put(key []byte, value []byte) {
	err := s.db.Put(key, value, nil)
	if err != nil {
		log.Infof("%s put key[%v] value[%v] thorw err[%v]", s, key, value, err)
	}
}

func (s *Storage) Has(key []byte) bool {
	ok, err := s.db.Has(key, nil)
	if err != nil {
		log.Infof("%s has key[%v] thorw err[%v]", s, key, err)
	}
	return ok
}

func (s *Storage) Delete(key []byte) {
	err := s.db.Delete(key, nil)
	if err != nil {
		log.Infof("%s delete key[%v] thorw err[%v]", s, key, err)
	}
}

func (s *Storage) Write(batch *leveldb.Batch) {
	err := s.db.Write(batch, nil)
	if err != nil {
		log.Infof("%s write batch[%v] thorw err[%v]", s, batch, err)
	}
}

func (s *Storage) Count() uint64 {
	var count uint64
	it := s.db.NewIterator(nil, nil)
	defer it.Release()
	for it.First(); it.Valid(); it.Next() {
		count++
	}
	return count
}

func (s *Storage) Close() {

	err := s.db.Close()
	if err != nil {
		log.Infof("%s close thorw err[%v]", s, err)
	}
}

func (s *Storage) String() string {
	return fmt.Sprintf("[id:%d,path:%s]", s.id, s.path)
}
