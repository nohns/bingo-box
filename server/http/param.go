package http

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (s *Server) requireParam(rw http.ResponseWriter, r *http.Request, param string) (string, bool) {
	val, ok := mux.Vars(r)[param]
	if !ok {
		s.writeJsonPayload(rw, http.StatusBadRequest, "Bad url parameters", map[string]string{"paramName": param})
		return "", false
	}

	return val, true
}
