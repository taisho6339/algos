package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	yt := `---
persistent_queue_limit: 1048576

filter_configs:
- name: keywords_filter
  regex_filter_config:
    expression: "^.*(Java|Golang|Go|Python|Rust).*$"

scrape_configs:
- job: feed_tech_news
  site_url: "https://example.test.com"
  position_file_dir: /tmp/positions/feed_tech_news
  filter_name: keywords_filter
  notification_name: slack
  content_element:
    id: id
    body: body
    urlTemplate: urlTemplate

notification_configs:
- name: slack
  slack_notification_config:
    api_url: "https://xxxx.com"
    channel: test
`
	expected := &Config{
		PersistentQueueLimit: 1048576,
		FilterConfigs: []FilterConfig{
			{
				Name: "keywords_filter",
				RegexFilterConfigs: RegexFilterConfig{
					Expression: "^.*(Java|Golang|Go|Python|Rust).*$",
				},
			},
		},
		ScrapeConfigs: []ScrapeConfig{
			{
				Job:              "feed_tech_news",
				PositionFilePath: "/tmp/positions/feed_tech_news",
				SiteURLTemplate:  "https://example.test.com",
				FilterName:       "keywords_filter",
				NotificationName: "slack",
				ContentElement: ContentElement{
					ID:   "id",
					Body: "body",
					URL:  "urlTemplate",
				},
			},
		},
		NotificationConfigs: []NotificationConfig{
			{
				Name: "slack",
				SlackNotificationConfig: SlackNotificationConfig{
					APIURL:  "https://xxxx.com",
					Channel: "test",
				},
			},
		},
	}
	config, err := Parse(yt)
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Exactly(t, expected, config)
}
