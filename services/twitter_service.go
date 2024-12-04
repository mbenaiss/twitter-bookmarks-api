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
    url := fmt.Sprintf("https://api.twitter.com/2/users/%s/bookmarks?tweet.fields=created_at,author_id,text&expansions=author_id&user.fields=username,name", userID)
    
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }

    req.Header.Add("Authorization", fmt.Sprintf("OAuth oauth_consumer_key=\"%s\", oauth_token=\"%s\"", s.config.TwitterAPIKey, s.config.TwitterAPISecret))
    req.Header.Add("Content-Type", "application/json")
    
    resp, err := s.client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("Twitter API error: %d", resp.StatusCode)
    }

    var twitterResp struct {
        Data []struct {
            ID        string    `json:"id"`
            Text      string    `json:"text"`
            CreatedAt time.Time `json:"created_at"`
            AuthorID  string    `json:"author_id"`
        } `json:"data"`
        Includes struct {
            Users []struct {
                ID       string `json:"id"`
                Username string `json:"username"`
                Name     string `json:"name"`
            } `json:"users"`
        } `json:"includes"`
        Meta struct {
            NextToken string `json:"next_token"`
        } `json:"meta"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&twitterResp); err != nil {
        return nil, err
    }

    // Create a map of user IDs to user data
    userMap := make(map[string]struct {
        Username string
        Name     string
    })
    for _, user := range twitterResp.Includes.Users {
        userMap[user.ID] = struct {
            Username string
            Name     string
        }{
            Username: user.Username,
            Name:     user.Name,
        }
    }

    bookmarks := make([]models.Bookmark, 0)
    for _, tweet := range twitterResp.Data {
        author := models.Author{
            ID: tweet.AuthorID,
        }
        if userData, ok := userMap[tweet.AuthorID]; ok {
            author.Username = userData.Username
            author.Name = userData.Name
        }
        
        bookmarks = append(bookmarks, models.Bookmark{
            ID:        tweet.ID,
            TweetID:   tweet.ID,
            Text:      tweet.Text,
            CreatedAt: tweet.CreatedAt,
            Author:    author,
        })
    }

    return &models.BookmarkResponse{
        Bookmarks: bookmarks,
        NextToken: twitterResp.Meta.NextToken,
    }, nil
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
