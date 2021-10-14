package internal

import "time"

type Item struct {
	Date     time.Time
	Title    string
	Category string
	URL      string
	ID       string
}
