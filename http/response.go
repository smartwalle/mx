package http

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func NewResponse(code int, data interface{}) *Response {
	var r = &Response{}
	r.Code = code
	r.Data = data
	return r
}

func (r *Response) Write(w http.ResponseWriter) error {
	return json.NewEncoder(w).Encode(r)
}
