# 🤖 hb-press-bot

This is running on a server and posting news from [Medienservice der Senatskanzlei](https://www.senatspressestelle.bremen.de/) ("Bremen Senate Chancellery Mediaservice") to the [Senatspressestelle inoffiziell](https://norden.social/web/accounts/107096555807402379#) Mastodon account.

# 🤨 Why?

The Mediaservice's webpage offers no RSS feed.

# 🤓 How?

This project is written in Go. The implementation is specific to the Mediaservice webpage mentioned above, but I guess the approach I chose can be used for any kind of webpage.

A loop runs in a configurable interval (currently every hour) and fetches the main webpage:

1. We fetch the raw HTML page and use [goquery](https://github.com/PuerkitoBio/goquery) to parse the table and scrap the titles and links from the page.
2. Every news item is stored in a simple file-based key-value database ([bbolt](https://go.etcd.io/bbolt)). The page ID is used as the key. All available metadata (title, URL, category) and the current timestamp are marshaled into a binary value using Go's gob encoding (it's not _really_ necessary to keep all the metadata except the title, but since we already have it ... 😉).
3. If the news item wasn't stored before _OR_ its title was updated, we publish via [go-mastodon](https://github.com/mattn/go-mastodon) to Mastodon. This is particular to this Mediaservice: sometimes there are "corrections" and previously published news items get republished.

Keeping all published news items in the database allows the code to be stateless and we also don't have to do any matching of already posted status updates on Mastodon.

# ⚙️ Configuration

This needs some configuration (mostly to make it debugable during development).

The basic configuration is provided by Go's flag package. Mastodon requires some more config, which will be read from a `mastodon.json` file.

Here is the file structure:

```json
{
  "Server": "https://mastodon.social",
  "ClientID": "CLIENT_ID",
  "ClientSecret": "CLIENT_SECRET_BASE64",
  "AccessToken": "ACCESS_TOKEN_BASE64"
}
```
