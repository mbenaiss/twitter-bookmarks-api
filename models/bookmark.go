package models

import "time"

type Bookmark struct {
    ID        string    `json:"id"`
    TweetID   string    `json:"tweet_id"`
    Text      string    `json:"text"`
    CreatedAt time.Time `json:"created_at"`
    Author    Author    `json:"author"`
}

type Author struct {
    ID       string `json:"id"`
    Username string `json:"username"`
    Name     string `json:"name"`
}

type BookmarkResponse struct {
    Bookmarks []Bookmark `json:"bookmarks"`
    NextToken string     `json:"next_token,omitempty"`
}
