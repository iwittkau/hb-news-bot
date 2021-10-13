package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
)

const (
	fetcherTypeStatic = "static"
	fetcherTypeHTTP   = "http"
)

type httpFetcher struct{}

func (*httpFetcher) Fetch(ctx context.Context, url string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating HTTP request")
	}
	req.Header.Set("User-Agent", userAgent)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending HTTP request: %w", err)
	}
	return res.Body, nil
}

type staticFetcher struct{}

func (*staticFetcher) Fetch(ctx context.Context, url string) (io.ReadCloser, error) {
	f, err := os.Open("testdata/page.html")
	if err != nil {
		return nil, err
	}
	return f, nil
}
