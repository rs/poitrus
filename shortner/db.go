package shortner

import (
	"github.com/dgraph-io/badger"
)

type DB struct {
	db *badger.DB
}

func NewDB(path string) (*DB, error) {
	opts := badger.DefaultOptions
	opts.Dir = path
	opts.ValueDir = path
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	return &DB{
		db: db,
	}, nil
}

func (s *DB) Close() error {
	return s.db.Close()
}

func (d *DB) Get(path string) (target string, found bool, err error) {
	err = d.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(path))
		if err != nil {
			return err
		}
		val, err := item.Value()
		if err != nil {
			return err
		}
		target = string(val)
		return nil
	})
	if err == badger.ErrKeyNotFound {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return target, true, nil
}

func (d *DB) Set(path, target string) (exists bool, err error) {
	err = d.db.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get([]byte(path)); err == nil {
			exists = true
			return nil
		} else if err != badger.ErrKeyNotFound {
			return err
		}
		err := txn.Set([]byte(path), []byte(target))
		return err
	})
	return
}

func (d *DB) Delete(path string) (found bool, err error) {
	err = d.db.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get([]byte(path)); err == nil {
			found = true
		} else if err == badger.ErrKeyNotFound {
			return nil
		} else {
			return err
		}
		return txn.Delete([]byte(path))
	})
	return
}
