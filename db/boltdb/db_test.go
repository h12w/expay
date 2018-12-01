package boltdb

import (
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"

	"h12.io/expay"
)

func TestNewFailed(t *testing.T) {
	dir, err := ioutil.TempDir(".", "test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	// create db file on an dir
	if _, err := New(dir); err == nil {
		t.Fatal("expect error but got nil")
	}
}

func TestBucketOps(t *testing.T) {
	dir, err := ioutil.TempDir(".", "test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	db, err := New(path.Join(dir, "db.bolt"))
	if err != nil {
		t.Fatal(err)
	}
	bucket := db.Bucket("test")

	// Create
	input1 := "abc"
	id1, err := bucket.Create(input1)
	if err != nil {
		t.Fatal(err)
	}

	// Get
	output1 := ""
	if err := bucket.Get(id1, &output1); err != nil {
		t.Fatal(err)
	}
	if output1 != input1 {
		t.Fatalf("expect %s got %s", input1, output1)
	}

	// Update
	input2 := "def"
	if err := bucket.Update(id1, input2); err != nil {
		t.Fatal(err)
	}

	input3 := "ghi"
	id3, err := bucket.Create(input3)
	if err != nil {
		t.Fatal(err)
	}

	it, err := bucket.List()
	if err != nil {
		t.Fatal(err)
	}
	ids := []string{}
	values := []string{}
	for it.Next() {
		value := ""
		id, err := it.Scan(&value)
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, id)
		values = append(values, value)
	}
	if err := it.Close(); err != nil {
		t.Fatal(err)
	}
	wantIDs := []string{id1, id3}
	if !reflect.DeepEqual(ids, wantIDs) {
		t.Fatalf("expect ids %v got %v", wantIDs, ids)
	}
	wantValues := []string{input2, input3}
	if !reflect.DeepEqual(values, wantValues) {
		t.Fatalf("expect values %v got %v", wantValues, values)
	}

	// Get 2
	output2 := ""
	if err := bucket.Get(id1, &output2); err != nil {
		t.Fatal(err)
	}
	if output2 != input2 {
		t.Fatalf("expect %s got %s", input2, output2)
	}

	// Delete
	if err := bucket.Delete(id1); err != nil {
		t.Fatal(err)
	}

	// Get deleted
	output3 := ""
	if err := bucket.Get(id1, &output3); err != expay.ErrNotFound {
		t.Fatalf("expect error %v got %v", expay.ErrNotFound, err)
	}
}
