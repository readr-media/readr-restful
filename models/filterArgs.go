package models

import (
	"time"
)

type FilterArgs struct {
	MaxResult   int                  `form:"max_result"`
	Page        int                  `form:"page"`
	Sorting     string               `form:"sort"`
	ID          int64                `form:"id"`
	Slug        string               `form:"slug"`
	Mail        string               `form:"mail"`
	Nickname    string               `form:"nickname"`
	Title       []string             `form:"title"`
	Description []string             `form:"description"`
	Content     []string             `form:"content"`
	Author      []string             `form:"author"`
	Tag         []string             `form:"tag"`
	PublishedAt map[string]time.Time `form:"published_at"`
	CreatedAt   map[string]time.Time `form:"created_at"`
	UpdatedAt   map[string]time.Time `form:"updated_at"`
}
