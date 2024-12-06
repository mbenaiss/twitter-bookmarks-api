package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
	"twitter-bookmarks/config"

	"github.com/gin-gonic/gin"
)

var (
	accessToken string
)

func main() {
	redirectURI := "https://v-price-guilty-proceeding.trycloudflare.com/callback"
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to get config: %v", err)
	}

	clientID := cfg.TwitterClientID

	router := gin.Default()
	codeVerifier := GenerateCodeVerifier()
	codeChallenge := GenerateCodeChallenge(codeVerifier)

	state := "my-state"
	// Route to generate auth URL and redirect user
	router.GET("/login", func(c *gin.Context) {
		authURL := fmt.Sprintf(
			"https://twitter.com/i/oauth2/authorize?response_type=code&client_id=%s&redirect_uri=%s&scope=bookmark.read tweet.read users.read&state=%s&code_challenge=%s&code_challenge_method=s256",
			clientID, url.QueryEscape(redirectURI), state, codeChallenge,
		)
		c.Redirect(http.StatusFound, authURL)
	})

	// Callback route to handle OAuth redirect
	router.GET("/callback", func(c *gin.Context) {
		queryState := c.Query("state")
		code := c.Query("code")

		if queryState != state {
			c.JSON(http.StatusBadRequest, gin.H{"error": "State mismatch"})
			return
		}

		// Exchange authorization code for access token
		data := url.Values{}
		data.Set("client_id", clientID)
		data.Set("grant_type", "authorization_code")
		data.Set("redirect_uri", redirectURI)
		data.Set("code", code)
		data.Set("code_verifier", codeVerifier)

		req, err := http.NewRequest("POST", "https://api.twitter.com/2/oauth2/token", strings.NewReader(data.Encode()))
		if err != nil {
			log.Println("Failed to create request:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Println("Failed to make request:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			log.Println("Error response from Twitter:", string(body))
			c.JSON(resp.StatusCode, gin.H{"error": "Twitter API error"})
			return
		}

		var tokenResponse map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
			log.Println("Failed to parse response:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		accessToken = tokenResponse["access_token"].(string)
		c.Redirect(http.StatusFound, "/tweets")
	})

	// Route to revoke the access token
	router.GET("/revoke", func(c *gin.Context) {
		req, err := http.NewRequest("POST", "https://api.twitter.com/oauth2/revoke", nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
			return
		}

		req.Header.Set("Authorization", "Bearer "+accessToken)
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Println("Failed to revoke token:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke token"})
			return
		}
		defer resp.Body.Close()

		c.JSON(http.StatusOK, gin.H{"message": "Token revoked"})
	})

	// Route to get bookmarks
	router.GET("/tweets", func(c *gin.Context) {
		req, err := http.NewRequest("GET", "https://api.twitter.com/2/users/me/bookmarks", nil)
		if err != nil {
			log.Println("Failed to create request:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		req.Header.Set("Authorization", "Bearer "+accessToken)
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Println("Failed to fetch bookmarks:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			log.Println("Error response from Twitter:", string(body))
			c.JSON(resp.StatusCode, gin.H{"error": "Twitter API error"})
			return
		}

		var bookmarksResponse map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&bookmarksResponse); err != nil {
			log.Println("Failed to parse response:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		c.JSON(http.StatusOK, bookmarksResponse)
	})

	log.Printf("Server is running at http://https://v-price-guilty-proceeding.trycloudflare.com             ")
	router.Run(":3030")
}

// GenerateCodeVerifier creates a code verifier for PKCE
func GenerateCodeVerifier() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._~"
	const length = 43

	verifier := make([]byte, length)
	rand.Seed(time.Now().UnixNano())
	for i := range verifier {
		verifier[i] = charset[rand.Intn(len(charset))]
	}
	return string(verifier)
}

// GenerateCodeChallenge creates a code challenge from the code verifier
func GenerateCodeChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}
