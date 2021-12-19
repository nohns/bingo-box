package http

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (s *Server) getGames() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func (s *Server) getGame() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func (s *Server) postGame() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func (s *Server) patchGame() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func (s *Server) deleteGame() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func (s *Server) getCardMatch() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func (s *Server) registerGameRoutes(r *mux.Router, middleware ...mux.MiddlewareFunc) {

	r.Use(middleware...)

	r.HandleFunc("/", s.getGames()).Methods(http.MethodGet)
	r.HandleFunc("/", s.postGame()).Methods(http.MethodPost)
	r.HandleFunc("/:gameId", s.getGame()).Methods(http.MethodGet)
	r.HandleFunc("/:gameId", s.patchGame()).Methods(http.MethodPatch)
	r.HandleFunc("/:gameId", s.deleteGame()).Methods(http.MethodDelete)

	// Actions methods
	r.HandleFunc("/:gameId/matchCard/:cardNumber", s.getCardMatch()).Methods(http.MethodGet)
}
