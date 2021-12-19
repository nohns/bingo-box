package http

import (
	"net"
	"net/http"

	"github.com/gorilla/mux"
	bingo "github.com/nohns/bingo-box/server"
	"github.com/nohns/bingo-box/server/logger"
)

type Server struct {
	http           *http.Server
	router         *mux.Router
	ln             net.Listener
	authMiddleware mux.MiddlewareFunc

	Log logger.Logger

	// Exposed dependencies
	Addr        string
	UserService *bingo.UserService
	GameService *bingo.GameService
}

// Start to listen on http address and serve http request. Block until error occurs.
func (s *Server) Serve() error {

	// Create listener
	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}

	// Set listener and serve http
	s.ln = ln
	return s.http.Serve(s.ln)
}

func (s *Server) serveHTTP(w http.ResponseWriter, r *http.Request) {

	// Pass handling to mux router
	s.router.ServeHTTP(w, r)
}

func NewServer(apiKey string) *Server {
	s := &Server{
		http:   &http.Server{},
		router: mux.NewRouter(),
	}

	// Setup designated handler and listener
	s.http.Handler = http.HandlerFunc(s.serveHTTP)

	// Create resource routers
	gameRtr := s.router.PathPrefix("/games").Subrouter()
	authRtr := s.router.PathPrefix("/auth").Subrouter()

	// Register shared middleware
	s.authMiddleware = apiKeyMiddlewareFactory(apiKey)

	// Register resource routes. Some with middleware
	s.registerGameRoutes(gameRtr, s.authMiddleware)
	s.registerAuthRoutes(authRtr)

	return s
}
