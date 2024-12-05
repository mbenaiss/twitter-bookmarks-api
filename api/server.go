package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"twitter-bookmarks/api/middleware"
)

type Options func(*Server)

// Server is a struct representing a http Server.
type Server struct {
	httpServer *http.Server
	handler    *gin.Engine
}

// New creates a new Server instance.
func New(port string, options ...Options) *Server {
	handler := gin.Default()
	handler.Use(middleware.Logger())
	handler.Use(middleware.CORS())

	s := &Server{
		httpServer: &http.Server{
			Addr: fmt.Sprintf("0.0.0.0:%s", port),
		},
		handler: handler,
	}

	for _, o := range options {
		o(s)
	}

	return s
}

// StartHTTP start the http server.
func (s *Server) StartHTTP() error {
	s.httpServer.Handler = s.handler

	return s.httpServer.ListenAndServe()
}

// Shutdown shutdown the http server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// handleWithRetry wraps a handler with retry logic for 401 errors
func (s *Server) handleWithRetry(handler gin.HandlerFunc, service service) gin.HandlerFunc {
	return func(c *gin.Context) {
		handler(c)
		
		// Si on reçoit une erreur 401, on essaie de récupérer un nouveau token et de réessayer
		if c.Writer.Status() == http.StatusUnauthorized {
			tokenMiddleware := middleware.GetTokenFromTwitter(service)
			tokenMiddleware(c)
			
			// Si l'authentification a réussi, on réessaie l'appel original
			if c.Writer.Status() != http.StatusUnauthorized {
				handler(c)
			}
		}
	}
}

// WithRegisterRoutes register the routes for the server.
func WithRegisterRoutes(service service, secretKey string) Options {
	return func(s *Server) {
		s.handler.Use(middleware.Auth(secretKey))

		s.handler.GET("/bookmarks", s.handleWithRetry(s.getBookmarks(service), service))
		s.handler.GET("/bookmarks/filter", s.handleWithRetry(s.getBookmarksWithDateFilter(service), service))
	}
}
