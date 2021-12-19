package http

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	bingo "github.com/nohns/bingo-box/server"
)

func (s *Server) getAuthenticate() http.HandlerFunc {
	type requestBody struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}
	return func(rw http.ResponseWriter, r *http.Request) {
		// Parse request json body
		var body requestBody
		if !s.jsonBody(rw, r, &body) {
			return
		}

		// Response payload. Defer writing payload to response writer until handler has returned
		var status int
		var message string
		var data interface{}

		user, err := s.UserService.Authenticate(r.Context(), body.Email, body.Password)
		if err != nil {
			s.Log.Errf("could not authenticate user due to error:\n%v\n", err)

			// Try to check what kind of error we are dealing withs
			switch {
			case
				errors.Is(err, bingo.ErrUserNotFound),
				errors.Is(err, bingo.ErrPasswordMismatch):
				status = http.StatusUnauthorized
				message = "Credentials did not match any users"
			default:
				status = http.StatusInternalServerError
				message = "Unknown error occured"
			}

			s.writeJsonPayload(rw, status, message, data)
			return
		}

		// Set response payload
		status = http.StatusOK
		data = user
		s.writeJsonPayload(rw, status, message, data)
	}
}

func (s *Server) postRefresh() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {

	}
}

func (s *Server) postRegister() http.HandlerFunc {
	type requestBody struct {
		Name     string `json:"name" validate:"required"`
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}

	return func(rw http.ResponseWriter, r *http.Request) {

		// Parse request json body
		var body requestBody
		if !s.jsonBody(rw, r, &body) {
			return
		}

		// Response payload. Defer writing payload to response writer until handler has returned
		var status int
		var message string
		var data interface{}

		user, err := s.UserService.Register(r.Context(), body.Name, body.Email, body.Password)
		if err != nil {
			s.Log.Errf("could not register user due to error:\n%v\n", err)

			// Try to check what kind of error we are dealing withs
			switch {
			case errors.Is(err, bingo.ErrUserAlreadyExists):
				status = http.StatusConflict
				data = validationData{
					"email": &validationField{
						FieldName: "email",
						Value:     body.Email,
						Errors: []validationFieldError{
							{
								Reason:  "unique",
								Message: "Email already in use",
							},
						},
					},
				}
			default:
				status = http.StatusInternalServerError
				message = "Unknown error occured"
			}

			s.writeJsonPayload(rw, status, message, data)
			return
		}

		// Set response payload
		status = http.StatusCreated
		data = user

		s.writeJsonPayload(rw, status, message, data)
	}
}

func (s *Server) registerAuthRoutes(r *mux.Router) {

	r.HandleFunc("/login", s.getAuthenticate()).Methods(http.MethodGet)
	r.HandleFunc("/register", s.postRegister()).Methods(http.MethodPost)
	r.HandleFunc("/refresh", s.postRefresh()).Methods(http.MethodPost)
}
