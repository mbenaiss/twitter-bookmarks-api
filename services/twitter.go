package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"twitter-bookmarks/models"
)

// TwitterService handles communication with the Twitter API
type TwitterService struct {
	client    *http.Client
	apiKey    string
	secretKey string
}

// NewTwitterService creates a new TwitterService
func NewTwitterService(apiKey, secretKey string) *TwitterService {
	return &TwitterService{
		client: &http.Client{
			Timeout: time.Second * 10,
		},
		apiKey:    apiKey,
		secretKey: secretKey,
	}
}

// Authenticate authenticates with the Twitter API
func (s *TwitterService) Authenticate(ctx context.Context) (string, error) {
	url := "https://api.twitter.com/oauth2/token"
	credentials := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", s.apiKey, s.secretKey)))
	data := strings.NewReader("grant_type=client_credentials")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, data)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", credentials))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Twitter API error: status=%d", resp.StatusCode)
	}

	var authResp struct {
		AccessToken string `json:"access_token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return authResp.AccessToken, nil
}

// GetBookmarks gets the bookmarks for a user
func (s *TwitterService) GetBookmarks(ctx context.Context, userID string, token string) (*models.BookmarkResponse, error) {
	url := fmt.Sprintf("https://api.twitter.com/2/users/%s/bookmarks?tweet.fields=created_at,author_id,text&expansions=author_id&user.fields=username,name", userID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

	var lastErr error
	maxRetries := 3
	
	for i := 0; i < maxRetries; i++ {
		resp, err := s.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusTooManyRequests {
			lastErr = fmt.Errorf("rate limit exceeded")
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("Twitter API error: status=%d", resp.StatusCode)
		}
		
		return s.parseBookmarksResponse(resp)
	}
	
	return nil, fmt.Errorf("failed after %d retries: %v", maxRetries, lastErr)
}

// GetBookmarksAfterDate gets the bookmarks for a user after a specific date
func (s *TwitterService) GetBookmarksAfterDate(ctx context.Context, userID string, after time.Time, token string) (*models.BookmarkResponse, error) {
	bookmarks, err := s.GetBookmarks(ctx, userID, token)
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

func (s *TwitterService) parseBookmarksResponse(resp *http.Response) (*models.BookmarkResponse, error) {
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
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

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
