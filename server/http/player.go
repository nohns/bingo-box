package http

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nohns/bingo-box/server/pdf"
)

func (s *Server) getCardsPdf() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {

		// Get player id from url
		playerId, ok := s.requireParam(rw, r, "playerID")
		if !ok {
			return
		}

		// Response payload
		var status int
		var message string
		var data interface{}

		player, err := s.PlayerService.Get(r.Context(), playerId)
		if err != nil {
			s.Log.Errf("could not find cards for player id %s due to error:\n%v\n", playerId, err)

			// Try to check what kind of error we are dealing with
			switch {
			default:
				status = http.StatusInternalServerError
				message = "Unknown error occured"
			}

			s.writeJsonPayload(rw, status, message, data)
			return
		}

		// Generate cards pdf
		cardsPdf, err := pdf.GenFromCards(player.Invitation.Game, player.Cards)
		if err != nil {
			status = http.StatusInternalServerError
			message = "Pdf file could not be generated"
			return
		}

		// Write pdf to response
		rw.Header().Set("Content-Type", "application/pdf")
		cardsPdf.Write(rw)
	}
}

func (s *Server) RegisterPlayerRoutes(r *mux.Router, mw ...mux.MiddlewareFunc) {
	r.Use(mw...)

	r.HandleFunc("/{playerID}/cards", s.getCardsPdf()).Methods(http.MethodGet)
}
