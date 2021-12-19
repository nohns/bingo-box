package http

import (
	"encoding/json"
	"errors"
	"net/http"
)

// Converts json request body into struct of target type.
// If unmarshalling failes, then send 400 Bad Request error.
// Returns a bool on whether the body was parsed successfully or not.
func (s *Server) jsonBody(rw http.ResponseWriter, r *http.Request, target interface{}) bool {

	// Check content-type
	if ct := r.Header.Get("Content-Type"); ct != "application/json" {
		s.writeJsonPayload(rw, http.StatusBadRequest, "Bad header Content-Type. Expected application/json", nil)
		s.Log.Errf("request received with bad content-type = %s\n", ct)
		return false
	}

	// Try to unmarshal body
	err := json.NewDecoder(r.Body).Decode(target)
	if err != nil {
		s.writeJsonPayload(rw, http.StatusBadRequest, "Malformed request body", nil)
		s.Log.Errf("could not unmarshal json request body:\n%v\n", err)
		return false
	}

	// Validate body
	if err, ok := validateBody(target); !ok {
		status := http.StatusInternalServerError
		message := "Unknown error occured"
		var data interface{}

		var valiData validationData
		if errors.As(err, &valiData) {
			status = http.StatusBadRequest
			message = "Validation failed"
			data = valiData
		}

		s.writeJsonPayload(rw, status, message, data)
		s.Log.Errf("could not validate json request body:\n%v\n", err)
		return false
	}

	return true
}
