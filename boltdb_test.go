package main

import (
	"path"
	"reflect"
	"testing"

	"github.com/iwittkau/hb-news-bot/internal"
)

func Test_db_SetPublished(t *testing.T) {
	dbName := path.Join(t.TempDir(), "db_"+t.Name())
	db, err := newDB(dbName)
	if err != nil {
		t.Error(err)
	}
	t.Cleanup(func() { db.Close() })

	i := internal.Item{ID: "1234", Title: "Title"}
	ok, err := db.SetPublished(i)
	if err != nil {
		t.Error(err)
	}
	if ok {
		t.Error("item should not be published")
	}

	ok, err = db.SetPublished(i)
	if err != nil {
		t.Error(err)
	}
	if !ok {
		t.Error("item should be published")
	}

	i.Title = "Updated title"
	ok, err = db.SetPublished(i)
	if err != nil {
		t.Error(err)
	}
	if ok {
		t.Error("item should not be published")
	}

	u, err := db.Get(i.ID)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(i, *u) {
		t.Errorf("got %q, want %q", i, *u)
	}
}
