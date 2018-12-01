package payment

import (
	"errors"
	"reflect"
	"sort"
	"strconv"
	"sync"

	"h12.io/expay"
)

type fakeDB struct {
	m  map[string]interface{}
	id int
	mu sync.RWMutex

	// injected errors
	getErr       error
	createErr    error
	updateErr    error
	deleteErr    error
	listErr      error
	iterScanErr  error
	iterCloseErr error
}

type kv struct {
	key   string
	value interface{}
}
type fakeIterator struct {
	kvs      []kv
	i        int
	scanErr  error
	closeErr error
}

func newFakeDB() *fakeDB {
	return &fakeDB{m: make(map[string]interface{})}
}

func (db *fakeDB) Create(v interface{}) (id string, err error) {
	if db.createErr != nil {
		return "", db.createErr
	}
	db.mu.Lock()
	defer db.mu.Unlock()
	db.id++
	id = strconv.Itoa(db.id)
	db.m[id] = v
	return id, nil
}

func (db *fakeDB) Get(id string, v interface{}) error {
	if db.getErr != nil {
		return db.getErr
	}
	db.mu.RLock()
	defer db.mu.RUnlock()
	dbv, ok := db.m[id]
	if !ok {
		return expay.ErrNotFound
	}
	return scanValue(dbv, v)
}

func scanValue(dbv, v interface{}) error {
	in := reflect.ValueOf(dbv)
	for in.Kind() == reflect.Ptr {
		in = in.Elem()
	}
	out := reflect.ValueOf(v)
	for out.Kind() == reflect.Ptr {
		out = out.Elem()
	}
	if out.Type() != in.Type() {
		return errors.New("incompatible type")
	}
	out.Set(in)
	return nil
}

func (db *fakeDB) Update(id string, v interface{}) error {
	if db.updateErr != nil {
		return db.updateErr
	}
	db.mu.Lock()
	defer db.mu.Unlock()
	if _, ok := db.m[id]; !ok {
		return expay.ErrNotFound
	}
	db.m[id] = v
	return nil
}

func (db *fakeDB) Delete(id string) error {
	if db.deleteErr != nil {
		return db.deleteErr
	}
	db.mu.Lock()
	defer db.mu.Unlock()
	delete(db.m, id)
	return nil
}

func (db *fakeDB) List() (expay.Iter, error) {
	if db.listErr != nil {
		return nil, db.listErr
	}
	db.mu.RLock()
	defer db.mu.RUnlock()
	keys := make([]string, 0, len(db.m))
	for key := range db.m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	kvs := make([]kv, 0, len(db.m))
	for _, key := range keys {
		value := db.m[key]
		kvs = append(kvs, kv{key: key, value: value})
	}
	return &fakeIterator{
		kvs:      kvs,
		i:        -1,
		scanErr:  db.iterScanErr,
		closeErr: db.iterCloseErr,
	}, nil
}

func (it *fakeIterator) Next() bool {
	defer func() {
		it.i++
	}()
	return it.i < len(it.kvs)-1
}

func (it *fakeIterator) Close() error {
	return it.closeErr
}

func (it *fakeIterator) Scan(v interface{}) (id string, err error) {
	if it.scanErr != nil {
		return "", it.scanErr
	}
	kv := it.kvs[it.i]
	return kv.key, scanValue(kv.value, v)
}

// TODO: implement cursor-based pagination
func (db *fakeDB) Paginate(lastCursor string, limit int) (expay.Iter, error) {
	return nil, errors.New("not implemented yet")
}
