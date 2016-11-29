// This middleware provides request access token verification functionality
//
// Example use case:
//
//  routes.Handle("/some/route", a.Dependencies.MWHandler.Handle([]rye.Handler{
//     rye.NewMiddlewareAccessToken(tokenHeaderName, []string{token1, token2}),
//     yourHandler,
// })).Methods("POST")

package rye

import (
	"errors"
	"fmt"
	"net/http"
)

type accessTokens struct {
	headerName string
	tokens     []string
}

func NewMiddlewareAccessToken(headerName string, tokens []string) func(rw http.ResponseWriter, req *http.Request) *Response {
	a := &accessTokens{
		headerName: headerName,
		tokens:     tokens,
	}
	return a.handle
}

func (a *accessTokens) handle(rw http.ResponseWriter, r *http.Request) *Response {
	token := r.Header.Get(a.headerName)

	if token == "" {
		return &Response{
			Err:        fmt.Errorf("No access token found; ensure you pass '%s' in header", a.headerName),
			StatusCode: http.StatusUnauthorized,
		}
	}

	if ok := stringListContains(a.tokens, token); !ok {
		return &Response{
			Err:        errors.New("Unauthorized request: invalid access token"),
			StatusCode: http.StatusUnauthorized,
		}
	}

	return nil
}

func stringListContains(stringList []string, element string) bool {
	for _, v := range stringList {
		if v == element {
			return true
		}
	}

	return false
}
