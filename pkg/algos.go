package pkg

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type Content struct {
	ID   string
	Body string
	URL  string
}

type ContentFilter interface {
	Filter(c Content) bool
}

type Algos struct {
	cfg Config
}

func NewAlgos(cfg Config) *Algos {
	return &Algos{
		cfg: cfg,
	}
}

func (a *Algos) Run() error {
	wg := &sync.WaitGroup{}

	// TODO: ScrapeConfigの数だけgoroutineを起動する

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	wg.Wait()
	return nil
}
