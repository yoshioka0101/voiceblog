package entity

import "time"

type ArticleShareTarget struct {
	ID          int64
	ArticleID   int64
	Provider    string
	ExternalID  string
	ExternalURL string
	PublishedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
