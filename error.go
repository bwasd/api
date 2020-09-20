package api

import (
	"encoding/json"
	"net/http"
)

type apiError struct {
	Code    int
	Message string
}

// apiError replies to the request with an api error and HTTP code. The caller
// should ensure no further writes are made to w.
func (e *apiError) Write(w http.ResponseWriter) error {
	w.WriteHeader(e.Code)
	type err struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	type wrap struct {
		Errors []err `json:"errors"`
	}
	return json.NewEncoder(w).Encode(wrap{
		Errors: []err{
			{
				Code:    e.Code,
				Message: e.Message,
			},
		},
	})
}
