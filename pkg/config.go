package pkg

import (
	"time"

	yaml "gopkg.in/yaml.v3"
)

// Scrape Configs

type ExtractType string

const (
	ExtractTypeAttr ExtractType = "attr"
	ExtractTypeText ExtractType = "text"
)

type ElementSelector struct {
	Selector    string      `yaml:"selector"`
	ExtractType ExtractType `yaml:"extract_type"`
	Attr        string      `yaml:"attr"`
}

type ContentSelector struct {
	ListItem ElementSelector `yaml:"list_item_selector"`
	ID       ElementSelector `yaml:"id_selector"`
	URL      ElementSelector `yaml:"url_selector"`
	Content  ElementSelector `yaml:"content_selector"`
}

type ScrapeConfig struct {
	Job              string `yaml:"job"`
	PositionFilePath string `yaml:"position_file_dir"`
	StartPageOffset  uint32 `yaml:"start_offset"`
	PageOffsetLimit  uint32 `yaml:"page_offset_limit"`

	SiteURLTemplate string          `yaml:"site_url_template"`
	ContentSelector ContentSelector `yaml:"content_selector"`
	ScrapeInterval  time.Duration   `yaml:"scrape_interval"`
}

// Filter Configs

type RegexFilterConfig struct {
	Expression string `yaml:"expression"`
}

type FilterConfig struct {
	Name               string            `yaml:"name"`
	RegexFilterConfigs RegexFilterConfig `yaml:"regex_filter_config"`
}

// Notification Configs

type SlackNotificationConfig struct {
	APIURL  string `yaml:"api_url"`
	Channel string `yaml:"channel"`
}

type NotificationConfig struct {
	Name                    string                  `yaml:"name"`
	SlackNotificationConfig SlackNotificationConfig `yaml:"slack_notification_config"`
}

type Config struct {
	PersistentQueueLimit int `yaml:"persistent_queue_limit"`

	FilterConfigs       []FilterConfig       `yaml:"filter_configs"`
	ScrapeConfigs       []ScrapeConfig       `yaml:"scrape_configs"`
	NotificationConfigs []NotificationConfig `yaml:"notification_configs"`
}

func Parse(text string) (*Config, error) {
	c := &Config{}
	if err := yaml.Unmarshal([]byte(text), c); err != nil {
		return nil, err
	}
	return c, nil
}
