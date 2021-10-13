package main

import (
	"context"
	"log"
)

const (
	publisherTypeLog = "log"
)

type logPublisher struct{}

func (p *logPublisher) Publish(_ context.Context, i Item) error {
	log.Printf("publishing: '%s'\n", i.Title)
	return nil
}

func (p *logPublisher) Skip(_ context.Context, i Item) {
	log.Printf("skipping: '%s'\n", i.Title)
}
