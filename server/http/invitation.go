package http

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	bingo "github.com/nohns/bingo-box/server"
)

func (s *Server) postInvitation() http.HandlerFunc {
	type requestBody struct {
		GameID         string `json:"gameId" validate:"required"`
		DeliveryMethod string `json:"deliveryMethod" validate:"required"`
		MaxCardAmount  int    `json:"maxCardAmount" validate:"required"`
		Active         bool   `json:"active" validate:"required"`
		Criteria       []struct {
			Field string `json:"field" validate:"required"`
			Kind  string `json:"kind" validate:"required"`
			Value string `json:"value" validate:"required"`
		}
	}
	return func(rw http.ResponseWriter, r *http.Request) {
		// Parse request json body
		var body requestBody
		if !s.jsonBody(rw, r, &body) {
			return
		}

		// Response payload
		var status int
		var message string
		var data interface{}

		criteria := make([]bingo.InvitationCriterion, 0, len(body.Criteria))
		for _, c := range body.Criteria {
			criteria = append(criteria, bingo.InvitationCriterion{
				Field: bingo.InvitationCriterionField(c.Field),
				Kind:  bingo.InvitationCriterionKind(c.Kind),
				Value: c.Value,
			})
		}

		inv, err := s.InvitationService.Create(r.Context(), body.GameID, bingo.InvitationDeliveryMethod(body.DeliveryMethod), body.MaxCardAmount, criteria)
		if err != nil {
			s.Log.Errf("could not create invitation for given game id %s due to error:\n%v\n", body.GameID, err)

			// Try to check what kind of error we are dealing withs
			var valErr bingo.ValidationErr
			switch {
			case errors.As(err, &valErr):
				status = http.StatusBadRequest
				message = "Validation failed"
				data = translateBingoValidationErr(valErr)
			default:
				status = http.StatusInternalServerError
				message = "Unknown error occured"
			}

			s.writeJsonPayload(rw, status, message, data)
			return
		}

		// Set response payload
		status = http.StatusCreated
		data = inv
		s.writeJsonPayload(rw, status, message, data)
	}
}

func (s *Server) getInvitation() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {

		// Response payload
		var status int
		var message string
		var data interface{}

		// Get invitation id from url
		invId, ok := s.requireParam(rw, r, "invID")
		if !ok {
			return
		}

		inv, err := s.InvitationService.Get(r.Context(), invId)
		if err != nil {
			s.Log.Errf("could not get invitation for given invitation id %s due to error:\n%v\n", invId, err)

			// Try to check what kind of error we are dealing withs
			switch {
			default:
				status = http.StatusInternalServerError
				message = "Unknown error occured"
			}

			s.writeJsonPayload(rw, status, message, data)
			return
		}

		// Set response payload
		status = http.StatusOK
		data = inv
		s.writeJsonPayload(rw, status, message, data)
	}
}

func (s *Server) patchDisableInvitation() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		// Get invitation id from url
		invId, ok := s.requireParam(rw, r, "invID")
		if !ok {
			return
		}

		// Response payload
		var status int
		var message string
		var data interface{}

		inv, err := s.InvitationService.Deactivate(r.Context(), invId)
		if err != nil {
			s.Log.Errf("could not disable invitation for given invitation id %s due to error:\n%v\n", invId, err)

			// Try to check what kind of error we are dealing with
			switch {
			default:
				status = http.StatusInternalServerError
				message = "Unknown error occured"
			}

			s.writeJsonPayload(rw, status, message, data)
			return
		}

		// Set response payload
		status = http.StatusOK
		data = inv
		s.writeJsonPayload(rw, status, message, data)
	}
}

func (s *Server) joinInvitation() http.HandlerFunc {
	type requestBody struct {
		Email      string `json:"email" validate:"required,email"`
		Name       string `json:"name" validate:"required"`
		CardAmount int    `json:"cardAmount" validate:"required"`
	}
	return func(rw http.ResponseWriter, r *http.Request) {
		invId, ok := s.requireParam(rw, r, "invID")
		if !ok {
			return
		}

		// Parse request json body
		var body requestBody
		if !s.jsonBody(rw, r, &body) {
			return
		}

		// Response payload
		var status int
		var message string
		var data interface{}

		player, err := s.InvitationService.Join(r.Context(), invId, body.Name, body.Email, body.CardAmount)
		if err != nil {
			s.Log.Errf("could not join invitation for given inv id %s due to error:\n%v\n", invId, err)

			// Try to check what kind of error we are dealing withs
			var valErr bingo.ValidationErr
			switch {
			case errors.As(err, &valErr):
				status = http.StatusBadRequest
				message = "Validation failed"
				data = translateBingoValidationErr(valErr)
			default:
				status = http.StatusInternalServerError
				message = "Unknown error occured"
			}

			s.writeJsonPayload(rw, status, message, data)
			return
		}

		// Set response payload
		status = http.StatusCreated
		data = player
		s.writeJsonPayload(rw, status, message, data)
	}
}

func (s *Server) registerInvitationRoutes(r *mux.Router, middleware ...mux.MiddlewareFunc) {
	r.Use(middleware...)

	unauthedRtr := r.PathPrefix("/").Subrouter()
	authedRtr := r.PathPrefix("/").Subrouter()
	authedRtr.Use(s.authMiddleware)

	authedRtr.HandleFunc("/", s.postInvitation()).Methods(http.MethodPost)
	unauthedRtr.HandleFunc("/{invID}/disable", s.patchDisableInvitation()).Methods(http.MethodPatch)

	unauthedRtr.HandleFunc("/{invID}", s.getInvitation()).Methods(http.MethodGet)
	unauthedRtr.HandleFunc("/{invID}/join", s.joinInvitation()).Methods(http.MethodGet)
}
