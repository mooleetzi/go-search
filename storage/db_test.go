package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb"
)

func TestDB(t *testing.T) {
	db, err := Create("./test.db")
	assert.Nil(t, err)
	assert.NotNil(t, db)
	assert.Equal(t, "[id:0,path:./test.db]", db.String())
	defer db.Close()

	batch := &leveldb.Batch{}

	batchMap := make(map[string]string, 3)
	batchMap["1"] = "1"
	batchMap["2"] = "2"
	batchMap["999"] = "3"

	for k, v := range batchMap {
		batch.Put([]byte(k), []byte(v))
	}

	db.Write(batch)

	count := db.Count()
	assert.Equal(t, uint64(3), count)

	for k, v := range batchMap {
		val, err := db.Get([]byte(k))
		assert.Nil(t, err)
		assert.Equal(t, val, []byte(v))
	}
	batch.Reset()

	for k := range batchMap {
		batch.Delete([]byte(k))
	}
	db.Write(batch)
	for k := range batchMap {
		value, err := db.Get([]byte(k))
		assert.NotNil(t, err)
		assert.Equal(t, []byte{}, value)
		ok := db.Has([]byte(k))
		assert.False(t, ok)
	}

	db.Put([]byte("1"), []byte("1"))
	value, err := db.Get([]byte("1"))
	assert.Nil(t, err)
	assert.Equal(t, value, []byte("1"))

	value, err = db.Get([]byte("888"))
	assert.NotNil(t, err)
	assert.Nil(t, value)

	has1 := db.Has([]byte("1"))
	assert.True(t, has1)
	has2 := db.Has([]byte("2"))
	assert.False(t, has2)

	db.Delete([]byte("1"))
	has1 = db.Has([]byte("1"))
	assert.False(t, has1)

	db.Delete([]byte("2"))

	count = db.Count()
	assert.Equal(t, uint64(0), count)

	db2, err := Create("./test2.db")
	assert.Nil(t, err)
	assert.Equal(t, "[id:1,path:./test2.db]", db2.String())
}
