package mastodon

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/iwittkau/hb-news-bot/internal"
	"github.com/mattn/go-mastodon"
)

var _ internal.Publisher = (*Client)(nil)

type Client struct {
	mastodon *mastodon.Client
}

type Config struct {
	Server       string
	ClientID     string
	ClientSecret []byte
	AccessToken  []byte
}

func LoadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("os: %w", err)
	}
	defer f.Close()
	var c Config
	if err := json.NewDecoder(f).Decode(&c); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}
	return &c, nil
}

func (c *Config) toMastodonConfig() *mastodon.Config {
	return &mastodon.Config{
		Server:       c.Server,
		ClientID:     c.ClientID,
		ClientSecret: string(c.ClientSecret),
		AccessToken:  string(c.AccessToken),
	}
}

func NewClient(conf *Config) (*Client, error) {
	m := mastodon.NewClient(conf.toMastodonConfig())
	if _, err := m.GetAccountCurrentUser(context.Background()); err != nil {
		return nil, fmt.Errorf("account setup: %w", err)
	}
	c := &Client{
		mastodon: m,
	}
	return c, nil
}

func (c *Client) Publish(ctx context.Context, i internal.Item) error {
	t := mastodon.Toot{
		Status: fmt.Sprintf("%s\n%s\nRessort: %s", i.Title, i.URL, i.Category),
	}
	if _, err := c.mastodon.PostStatus(ctx, &t); err != nil {
		return fmt.Errorf("posting mastodon status: %w", err)
	}
	return nil
}

func (c *Client) Skip(context.Context, internal.Item) {}
