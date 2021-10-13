package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestDownload(t *testing.T) {
	t.Skip()
	res, err := http.Get(baseURL + newsPath)
	if err != nil {
		t.Error(err)
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Error()
	}
	if err := os.WriteFile("page.html", data, 0666); err != nil {
		t.Error(err)
	}
}

func Test_extractItems(t *testing.T) {
	f, err := os.Open("testdata/page.html")
	if err != nil {
		t.Errorf("opening cached page: %v", err)
	}
	extracted, err := extractItems(context.Background(), f)
	if err != nil {
		t.Errorf("extracting items: %v", err)
	}

	golden := readGoldenItems(t)

	for i, a := range extracted {
		a.Date = time.Time{}
		extracted[i] = a
	}
	for i, a := range golden {
		a.Date = time.Time{}
		golden[i] = a
	}
	if !reflect.DeepEqual(extracted, golden) {
		t.Errorf("got %q, want %q", golden, extracted)
	}
}

func readGoldenItems(t *testing.T) []Item {
	t.Helper()
	data, err := os.ReadFile("testdata/items_golden.json")
	if err != nil {
		t.Fatalf("reading golden items: %v", err)
	}
	var golden []Item
	if err := json.Unmarshal(data, &golden); err != nil {
		t.Fatalf("unmarshaling golden items: %v", err)
	}
	return golden
}

type testPublisher struct {
	sync.Mutex

	Published []Item
	Skipped   []Item
}

func (p *testPublisher) Publish(ctx context.Context, i Item) error {
	p.Lock()
	defer p.Unlock()
	p.Published = append(p.Published, i)
	return nil
}

func (p *testPublisher) Skip(ctx context.Context, i Item) {
	p.Lock()
	defer p.Unlock()
	p.Skipped = append(p.Skipped, i)
}

func Test_publish(t *testing.T) {
	dbName := path.Join(t.TempDir(), "db_"+t.Name())
	db, err := newDB(dbName)
	if err != nil {
		t.Error(err)
	}
	t.Cleanup(func() { db.Close() })

	golden := readGoldenItems(t)

	pub := &testPublisher{}
	s := aggregate{pub: pub, db: db}
	if err := s.publish(context.Background(), golden); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(pub.Published, golden) {
		t.Errorf("got %q, want %q", pub.Published, golden)
	}
}
