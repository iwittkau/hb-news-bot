package main

import (
	"context"
	"log"

	"github.com/iwittkau/hb-news-bot/internal"
)

const (
	publisherTypeLog = "log"
)

type logPublisher struct{}

func (p *logPublisher) Publish(_ context.Context, i internal.Item) error {
	log.Printf("publishing: '%s'\n", i.Title)
	return nil
}

func (p *logPublisher) Skip(_ context.Context, i internal.Item) {
	log.Printf("skipping: '%s'\n", i.Title)
}
