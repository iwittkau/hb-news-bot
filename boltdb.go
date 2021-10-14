package main

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"

	"github.com/iwittkau/hb-news-bot/internal"
	bolt "go.etcd.io/bbolt"
)

const bucketPostStatus = "posted"

type db struct {
	bolt *bolt.DB
}

func newDB(path string) (*db, error) {
	b, err := bolt.Open(path, 0666, nil)
	if err != nil {
		return nil, fmt.Errorf("opening bolt db: %w", err)
	}
	b.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte(bucketPostStatus)); err != nil {
			return fmt.Errorf("creating bucket: %w", err)
		}
		return nil
	})
	return &db{bolt: b}, nil
}

var errItemNotFound = errors.New("item not found")

func (db *db) Get(id string) (*internal.Item, error) {
	var i *internal.Item
	err := db.bolt.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketPostStatus))
		var err error
		i, err = viewItem(id, tx, b)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("getting item(%s): %w", id, err)
	}
	return i, nil
}

func viewItem(id string, tx *bolt.Tx, b *bolt.Bucket) (*internal.Item, error) {
	var i internal.Item
	data := b.Get([]byte(id))
	if data == nil {
		return nil, errItemNotFound
	}
	buf := bytes.NewBuffer(data)
	if err := gob.NewDecoder(buf).Decode(&i); err != nil {
		return nil, err
	}
	return &i, nil
}

func (db *db) SetPublished(i internal.Item) (bool /*previously published*/, error) {
	var posted bool
	err := db.bolt.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketPostStatus))
		existing, err := viewItem(i.ID, tx, b)
		switch {
		case errors.Is(err, errItemNotFound):
		case err != nil:
			return err
		}
		if existing != nil && existing.Title == i.Title {
			posted = true
			return nil
		}
		var buf bytes.Buffer
		if err := gob.NewEncoder(&buf).Encode(i); err != nil {
			return fmt.Errorf("encoding: %w", err)
		}
		if err := b.Put([]byte(i.ID), buf.Bytes()); err != nil {
			return fmt.Errorf("putting: %w", err)
		}
		return nil
	})
	return posted, err
}

func (db *db) Close() error {
	return db.bolt.Close()
}
