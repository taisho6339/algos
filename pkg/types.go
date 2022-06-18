package pkg

import (
	"sync"

	"golang.org/x/net/context"
)

// Scrapeして登録されたSubscriptionのchannelにEventを飛ばす
type Scraper interface {
	Start(ctx context.Context, wg *sync.WaitGroup, subscriptions []chan Event)
}

type Event struct {
	ID      string
	Content string
	URL     string
}
