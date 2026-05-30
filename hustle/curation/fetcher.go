package curation

import (
	"context"
	"github.com/mmcdole/gofeed"
	"time"
)

type FeedFetcher interface {
	Fetch(ctx context.Context, url string) ([]FeedItem, error)
}

type FeedItem struct {
	Title       string
	Description string
	Link        string
	Published   time.Time
	Content     string
}

type RSSFetcher struct {
	parser *gofeed.Parser
}

func NewRSSFetcher() *RSSFetcher {
	return &RSSFetcher{
		parser: gofeed.NewParser(),
	}
}

func (r *RSSFetcher) Fetch(ctx context.Context, url string) ([]FeedItem, error) {
	feed, err := r.parser.ParseURLWithContext(url, ctx)
	if err != nil {
		return nil, err
	}

	var items []FeedItem
	for _, item := range feed.Items {
		published := time.Now()
		if item.PublishedParsed != nil {
			published = *item.PublishedParsed
		}

		items = append(items, FeedItem{
			Title:       item.Title,
			Description: item.Description,
			Link:        item.Link,
			Published:   published,
			Content:     item.Content,
		})
	}

	return items, nil
}
