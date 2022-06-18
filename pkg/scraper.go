package pkg

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
	"github.com/saintfish/chardet"
	"github.com/taisho6339/algos/pkg/positions"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"golang.org/x/net/html/charset"
)

type htmlScraper struct {
	logger   *zap.Logger
	client   *http.Client
	position *positions.Position

	urlTemplate     string
	startPageOffset uint32
	pageOffsetLimit uint32
	contentSelector ContentSelector

	interval time.Duration
}

func NewHTMLScraper(logger *zap.Logger, config ScrapeConfig) Scraper {
	return &htmlScraper{
		logger:   logger,
		client:   http.DefaultClient,
		position: positions.NewPosition(config.PositionFilePath),
		interval: config.ScrapeInterval,

		urlTemplate:     config.SiteURLTemplate,
		startPageOffset: config.StartPageOffset,
		pageOffsetLimit: config.PageOffsetLimit,
		contentSelector: config.ContentSelector,
	}
}

func (s *htmlScraper) Start(ctx context.Context, wg *sync.WaitGroup, subscriptions []chan Event) {
	ticker := time.NewTicker(s.interval)
	ctx, cancel := context.WithCancel(ctx)
	defer ticker.Stop()
	defer cancel()
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("stopping...")
			return
		case <-ticker.C:
			ret, err := s.scrape(ctx)
			if err != nil {
				s.logger.Error("failed to scrape, stopping...", zap.String("reason", err.Error()))
				return
			}
			for i := range ret {
				fmt.Println("===")
				//fmt.Println(ret[i].ID)
				//fmt.Println(ret[i].Body)
				fmt.Println(ret[i].URL)
				fmt.Println("===")
			}
		}
	}
}

func (s *htmlScraper) scrape(ctx context.Context) ([]*Content, error) {
	lastID := s.position.ReadLastID()

	var results []*Content
	var newLastID string
	for i := s.startPageOffset; i < s.pageOffsetLimit+1; i++ {
		// get request
		url := fmt.Sprintf(s.urlTemplate, i)
		s.logger.Info("start scraping %s", zap.Any("url", url))
		doc, retryable, err := s.doRequest(ctx, url)
		if err != nil {
			s.logger.Error("failed to scrape", zap.Error(err))
			if retryable {
				return nil, nil
			}
			return nil, errors.Wrap(err, "not retryable error happens")
		}

		// parse html
		ret, hitThreshold := s.parseContents(doc, lastID)
		if len(ret) == 0 {
			if !hitThreshold {
				return nil, errors.New("seems to have not valid scrape config so that it can't scrape correctly")
			}
			s.logger.Info("there is no new content")
			break
		}
		results = append(results, ret...)
		if i == s.startPageOffset && len(ret) > 0 {
			newLastID = ret[0].ID
		}
		if hitThreshold {
			break
		}
	}
	if err := s.position.Save(newLastID); err != nil {
		s.logger.Error("failed to save position file", zap.String("reason", err.Error()))
	}
	return results, nil
}

func (s *htmlScraper) doRequest(ctx context.Context, url string) (doc *goquery.Document, retryable bool, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, false, errors.Wrap(err, "failed to request")
	}
	defer resp.Body.Close()

	// validate response
	if resp.StatusCode >= 300 {
		if resp.StatusCode >= 500 {
			return nil, true, errors.Errorf("failed to get content from urlTemplate, invalid status: %d", resp.StatusCode)
		}
		return nil, false, errors.Errorf("failed to get content from urlTemplate, invalid status: %d", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		return nil, false, errors.Errorf("invalid content type: %s", ct)
	}

	document, err := s.parseDocumentFromResponse(resp.Body)
	if err != nil {
		return nil, false, errors.Wrap(err, "failed to parse document from response")
	}
	return document, false, nil
}

func (s *htmlScraper) parseDocumentFromResponse(body io.Reader) (*goquery.Document, error) {
	buffer, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, err
	}

	detector := chardet.NewTextDetector()
	detectResult, err := detector.DetectBest(buffer)
	if err != nil {
		return nil, err
	}

	bufferReader := bytes.NewReader(buffer)
	reader, err := charset.NewReaderLabel(detectResult.Charset, bufferReader)
	if err != nil {
		return nil, err
	}
	return goquery.NewDocumentFromReader(reader)
}

var (
	re = regexp.MustCompile(`^(?:https?:\/\/)?(?:[^@\/\n]+@)?(?:www\.)?([^:\/\n]+)`)
)

func (s *htmlScraper) parseContents(doc *goquery.Document, thresholdID string) (ret []*Content, hitThreshold bool) {
	items := doc.Find(s.contentSelector.ListItem.Selector)
	items.Each(func(i int, selection *goquery.Selection) {
		if hitThreshold {
			return
		}
		var (
			id      string
			content string
			url     string
		)
		// id
		is := selection.Find(s.contentSelector.ID.Selector)
		if is == nil {
			return
		}
		if s.contentSelector.ID.ExtractType == ExtractTypeAttr {
			id, _ = is.Attr(s.contentSelector.ID.Attr)
		} else {
			id = is.Text()
		}
		if id == "" {
			s.logger.Warn("failed to extract id")
			return
		}
		if thresholdID == id {
			hitThreshold = true
			return
		}

		// content
		content = selection.Find(s.contentSelector.Content.Selector).Text()
		content = strings.TrimRight(strings.TrimRight(strings.TrimLeft(strings.TrimLeft(content, "\n"), "\t"), "\t"), "\n")
		if content == "" {
			s.logger.Warn("failed to extract content", zap.String("id", id))
			return
		}

		// url
		us := selection.Find(s.contentSelector.URL.Selector)
		if s.contentSelector.URL.ExtractType == ExtractTypeAttr {
			url, _ = us.Attr(s.contentSelector.URL.Attr)
		} else {
			url = us.Text()
		}
		if url == "" {
			s.logger.Warn("failed to extract url", zap.String("id", id))
		}
		if !strings.HasPrefix(url, "http") {
			domain := re.Find([]byte(s.urlTemplate))
			if domain != nil {
				url = fmt.Sprintf("%s%s", domain, url)
			}
		}

		ret = append(ret, &Content{
			ID:   id,
			Body: content,
			URL:  url,
		})
	})
	return
}
