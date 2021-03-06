package model

import (
	"time"
)

const (
	DiaryContentKey  = "content"
	DiaryImageKey    = "image"
	DiaryUserIDKey   = "user_id"
	DiaryDateKey     = "date"
	DiaryEmotionsKey = "emotions"
	DiaryThemeKey    = "theme"
)

type Diary struct {
	Content  string
	Image    string
	UserID   string `bson:"user_id"`
	Date     time.Time
	Emotions []string
	Theme    string
}
