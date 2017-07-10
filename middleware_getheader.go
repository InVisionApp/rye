package rye

import (
	"context"
	"net/http"
)

type getHeader struct {
	headerName string
	contextKey string
}

/*
NewMiddlewareGetHeader creates a new handler to extract any header and save its value into the context.
	headerName: the name of the header you want to extract
	contextKey: the value key that you would like to store this header under in the context

Example usage:

	routes.Handle("/some/route", a.Dependencies.MWHandler.Handle(
		[]rye.Handler{
			rye.NewMiddlewareGetHeader(headerName, contextKey),
			yourHandler,
		})).Methods("POST")
*/
func NewMiddlewareGetHeader(headerName, contextKey string) func(rw http.ResponseWriter, req *http.Request) *Response {
	h := getHeader{headerName: headerName, contextKey: contextKey}
	return h.getHeaderMiddleware
}

func (h *getHeader) getHeaderMiddleware(rw http.ResponseWriter, r *http.Request) *Response {
	rID := r.Header.Get(h.headerName)
	if rID != "" {
		return &Response{
			Context: context.WithValue(r.Context(), h.contextKey, rID),
		}
	}

	return nil
}
