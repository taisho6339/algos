package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/taisho6339/algos/pkg"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	config := pkg.ScrapeConfig{
		Job:              "askul",
		PositionFilePath: "/tmp/algos/askul.pos",
		StartPageOffset:  1,
		PageOffsetLimit:  3,
		SiteURLTemplate:  "https://www.askul.co.jp/m/14-0107-0107010-%d/?cateId=2&categoryM=0107010&categoryL=0107&categoryLl=14&deliveryEstimate=0&minPriceRange=&maxPriceRange=&sortDir=0&resultCount=50&resultType=0&variation=0&lstSelSpecCd=&priceRangeInputFlg=0&exclusionFlg=0&spcialDelivPicExclusionFlg=0&page=4",
		ContentSelector: pkg.ContentSelector{
			ListItem: pkg.ElementSelector{
				Selector: ".src1811__list__bd__ls__itm",
			},
			ID: pkg.ElementSelector{
				Selector:    ".src1811__tbl__info__nm > a",
				ExtractType: pkg.ExtractTypeAttr,
				Attr:        "href",
			},
			Content: pkg.ElementSelector{
				Selector: ".src1811__tbl__info__in > .src1811__tbl__info__nm",
			},
			URL: pkg.ElementSelector{
				Selector:    ".src1811__tbl__info__nm > a",
				ExtractType: pkg.ExtractTypeAttr,
				Attr:        "href",
			},
		},
		ScrapeInterval: time.Second * 10,
	}

	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	scraper := pkg.NewHTMLScraper(logger.With(zap.String("daemon", "scraper")), config)
	wg.Add(1)
	go scraper.Start(ctx, wg, nil)

	// handle signals
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		logger.Info("stopping daemons...")
		done <- true
	}()
	<-done
	cancel()
	wg.Wait()
}
