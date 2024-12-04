package services

import (
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    "twitter-bookmarks/config"
    "twitter-bookmarks/models"
)

type TwitterService struct {
    client  *http.Client
    config  *config.Config
}

func NewTwitterService() *TwitterService {
    return &TwitterService{
        client: &http.Client{
            Timeout: time.Second * 10,
        },
        config: config.Load(),
    }
}

func (s *TwitterService) GetBookmarks(userID string) (*models.BookmarkResponse, error) {
    url := fmt.Sprintf("https://api.twitter.com/2/users/%s/bookmarks", userID)
    
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }

    req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.config.TwitterAPIKey))
    
    resp, err := s.client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var bookmarkResp models.BookmarkResponse
    if err := json.NewDecoder(resp.Body).Decode(&bookmarkResp); err != nil {
        return nil, err
    }

    return &bookmarkResp, nil
}

func (s *TwitterService) GetBookmarksAfterDate(userID string, after time.Time) (*models.BookmarkResponse, error) {
    bookmarks, err := s.GetBookmarks(userID)
    if err != nil {
        return nil, err
    }

    var filteredBookmarks []models.Bookmark
    for _, bookmark := range bookmarks.Bookmarks {
        if bookmark.CreatedAt.After(after) {
            filteredBookmarks = append(filteredBookmarks, bookmark)
        }
    }

    return &models.BookmarkResponse{
        Bookmarks: filteredBookmarks,
        NextToken: bookmarks.NextToken,
    }, nil
}
