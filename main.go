package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"os/signal"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	baseURL  = "https://www.senatspressestelle.bremen.de/"
	newsPath = "pressemitteilungen-1464"

	userAgent = "Mastodon Bot"

	tableSelector = "table.bildboxtable tbody"

	dateFormat = "02.01.2006"
)

var spacesExpr = regexp.MustCompile(`\s+`)

type Item struct {
	Date     time.Time
	Title    string
	Category string
	URL      string
	ID       string
}

type publisher interface {
	Publish(context.Context, Item) error
	Skip(context.Context, Item)
}

type fetcher interface {
	Fetch(ctx context.Context, url string) (io.ReadCloser, error)
}

type aggregate struct {
	pub   publisher
	fetch fetcher
	db    *db
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	var (
		intv      string
		dbPath    string
		pubType   string
		fetchType string
	)
	flag.StringVar(&intv, "interval", "1h", "set interval for checking the source")
	flag.StringVar(&dbPath, "db", "bolt.db", "set path to db file")
	flag.StringVar(&pubType, "publisher", "log", `set publisher type ["log"]`)
	flag.StringVar(&fetchType, "fetcher", "static", `set fetcher type ["static", "http"]`)
	flag.Parse()
	var (
		d   time.Duration
		err error
	)
	if d, err = time.ParseDuration(intv); err != nil {
		return fmt.Errorf("parsing interval: %w", err)
	}

	var pub publisher
	switch pubType {
	case publisherTypeLog:
		pub = &logPublisher{}
	default:
		return fmt.Errorf("unknown publisher type: %q", pubType)
	}

	var fetch fetcher
	switch fetchType {
	case fetcherTypeStatic:
		fetch = &staticFetcher{}
	case fetcherTypeHTTP:
		fetch = &httpFetcher{}
	default:
		return fmt.Errorf("unknown fetcher type: %q", fetchType)
	}

	db, err := newDB(dbPath)
	if err != nil {
		return fmt.Errorf("creating db: %w", err)
	}
	defer db.Close()

	s := aggregate{
		pub:   pub,
		fetch: fetch,
		db:    db,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	start := time.Now().Round(d)
	if start.Before(time.Now()) {
		start = start.Add(d)
	}
	log.Printf("starting @%s, using publisher: %s, fetcher: %s\n", start, reflect.TypeOf(pub), reflect.TypeOf(fetch))
	select {
	case <-ctx.Done():
		return nil
	case <-time.After(time.Until(start)):
	}
	ticker := time.NewTicker(d)
	for {
		if err := s.run(ctx); err != nil {
			log.Println(err)
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

func (a *aggregate) run(ctx context.Context) error {
	r, err := a.fetch.Fetch(ctx, baseURL+newsPath)
	if err != nil {
		return fmt.Errorf("fetching page: %w", err)
	}
	defer r.Close()
	items, err := extractItems(ctx, r)
	if err != nil {
		return fmt.Errorf("extracting items: %w", err)
	}
	if err := a.publish(ctx, items); err != nil {
		return fmt.Errorf("publishing items: %w", err)
	}
	return nil
}

func (a *aggregate) publish(ctx context.Context, items []Item) error {
	for _, i := range items {
		published, err := a.db.SetPublished(i)
		if err != nil {
			return fmt.Errorf("db: %w", err)
		}
		if published {
			a.pub.Skip(ctx, i)
			continue
		}
		if err := a.pub.Publish(ctx, i); err != nil {
			return fmt.Errorf("publisher: %w", err)
		}
	}
	return nil
}

func extractItems(ctx context.Context, r io.Reader) ([]Item, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, fmt.Errorf("creating document: %w", err)
	}
	var res []Item
	table := doc.Find(tableSelector)
	table.Children().EachWithBreak(func(_ int, s *goquery.Selection) bool {
		i := Item{}
		s.Children().EachWithBreak(func(_ int, s *goquery.Selection) bool {
			if err = itemExtraction(s, &i); err != nil {
				return false
			}
			return ctx.Err() == nil
		})
		if err != nil {
			return false
		}
		if i.ID != "" {
			res = append(res, i)
		}
		return ctx.Err() == nil
	})
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return res, err
}

func itemExtraction(s *goquery.Selection, i *Item) error {
	v, ok := s.Attr("headers")
	if !ok {
		return nil
	}
	switch v {
	case "t1":
		var err error
		i.Date, err = time.Parse(dateFormat, s.Text())
		if err != nil {
			return fmt.Errorf("parsing date: %w", err)
		}
		i.Date = i.Date.Add(time.Hour * time.Duration(time.Now().Hour()))
	case "t2":
		page, ok := s.Children().First().Attr("href")
		if !ok {
			res, _ := s.Html()
			if res == "" {
				res = s.Text()
			}
			return fmt.Errorf("no link found in %s", res)
		}
		i.URL = baseURL + page
		i.Title = spacesExpr.ReplaceAllString(strings.TrimSpace(s.Text()), " ")
		q, err := url.ParseQuery(strings.TrimPrefix(page, "detail.php?"))
		if err != nil {
			return fmt.Errorf("unable to parse gsid: %w", err)
		}
		i.ID = q.Get("gsid")
	case "t3":
		i.Category = spacesExpr.ReplaceAllString(strings.TrimSpace(s.Text()), " ")
	}
	return nil
}
