package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"twitter-bookmarks/models"
)

type TwitterService struct {
	clientID     string
	clientSecret string
	redirectURI  string
	codeVerifier string
	refreshToken string
	client       *http.Client
	userID       string
}

func NewTwitterService(clientID, clientSecret, redirectURI string) *TwitterService {
	return &TwitterService{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURI:  redirectURI,
		client: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

// GenerateCodeVerifier creates a PKCE code verifier
func GenerateCodeVerifier() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._~"
	const length = 43
	verifier := make([]byte, length)
	for i := range verifier {
		verifier[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(verifier)
}

// Authenticate authenticates the user and retrieves an access token
func (s *TwitterService) Authenticate(ctx context.Context, authorizationCode string) (string, error) {
	s.codeVerifier = GenerateCodeVerifier()

	apiURL := "https://api.twitter.com/2/oauth2/token"

	data := url.Values{}
	data.Set("client_id", s.clientID)
	data.Set("grant_type", "authorization_code")
	data.Set("code", authorizationCode)
	data.Set("redirect_uri", s.redirectURI)
	data.Set("code_verifier", s.codeVerifier)

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("error from API: status=%d, body=%s", resp.StatusCode, string(body))
	}

	var tokenResponse struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		TokenType    string `json:"token_type"`
		Scope        string `json:"scope"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	s.refreshToken = tokenResponse.RefreshToken

	return tokenResponse.AccessToken, nil
}

// RefreshAccessToken renews the access token using the refresh token
func (s *TwitterService) RefreshAccessToken(ctx context.Context) (string, error) {
	if s.refreshToken == "" {
		return "", fmt.Errorf("no refresh token available")
	}

	apiURL := "https://api.twitter.com/2/oauth2/token"

	data := url.Values{}
	data.Set("client_id", s.clientID)
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", s.refreshToken)

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("error from API: status=%d, body=%s", resp.StatusCode, string(body))
	}

	var tokenResponse struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		TokenType    string `json:"token_type"`
		Scope        string `json:"scope"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if tokenResponse.RefreshToken != "" {
		s.refreshToken = tokenResponse.RefreshToken
	}

	return tokenResponse.AccessToken, nil
}

// GetBookmarks gets the bookmarks for a user
func (s *TwitterService) GetBookmarks(ctx context.Context, token string) (*models.BookmarkResponse, error) {
	url := fmt.Sprintf("https://api.twitter.com/2/users/%s/bookmarks", s.userID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("échec de création de la requête: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Twitter API error: status=%d", resp.StatusCode)
	}

	return s.parseBookmarksResponse(resp)
}

// GetBookmarksAfterDate gets the bookmarks for a user after a specific date
func (s *TwitterService) GetBookmarksAfterDate(ctx context.Context, token string, after time.Time) (*models.BookmarkResponse, error) {
	bookmarks, err := s.GetBookmarks(ctx, token)
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
