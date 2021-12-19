package http

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type httpPayload struct {
	Status  int         `json:"statusCode"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// Write json payload with http status code and error message from error interface.
func (s *Server) writeJsonFromError(rw http.ResponseWriter, status int, err error) {
	s.writeJsonPayload(rw, status, err.Error(), nil)
}

// Write json payload (message and data) with http status code
func (s *Server) writeJsonPayload(rw http.ResponseWriter, status int, message string, data interface{}) {

	// If a zero status code is given then something is not implemented correctly
	if status == 0 {
		s.Log.Errf("zero status code given. a handler not implemented properly")
		status = http.StatusNotImplemented
		message = "HTTP handler response not implemented"
	}

	httpErr := httpPayload{
		Status:  status,
		Message: message,
		Data:    data,
	}

	// Write payload to response writer
	s.writeJson(rw, status, httpErr)
}

// Writes json response with http status code and body to be marshalled into json.
func (s *Server) writeJson(rw http.ResponseWriter, status int, body interface{}) {

	// Try to marshall json response
	data, err := json.Marshal(body)
	if err != nil {
		s.Log.Errf("could not marshal response body to json:\n%v\n", err)
		rw.WriteHeader(status)
		return
	}

	// Set headers
	rw.Header().Set("Content-Type", "application/json")
	rw.Header().Set("Content-Length", strconv.Itoa(len(data)))

	// Write rson response
	rw.WriteHeader(status)
	_, err = rw.Write(data)
	if err != nil {
		s.Log.Errf("could not write json response body:\n%v\n", err)
		return
	}
}
