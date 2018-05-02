package rye

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
)

/*
NewMiddlewareAuth creates a new middleware to extract the Authorization header
from a request and validate it. It accepts a func of type AuthFunc which is
used to do the credential validation.
An AuthFuncs for Basic auth and JWT are provided here.

Example usage:

	routes.Handle("/some/route", myMWHandler.Handle(
		[]rye.Handler{
			rye.NewMiddlewareAuth(rye.NewBasicAuthFunc(map[string]string{
				"user1": "my_password",
			})),
			yourHandler,
		})).Methods("POST")
*/

type AuthFunc func(context.Context, string) *Response

func NewMiddlewareAuth(authFunc AuthFunc) func(rw http.ResponseWriter, req *http.Request) *Response {
	return func(rw http.ResponseWriter, r *http.Request) *Response {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			return &Response{
				Err:        errors.New("unauthorized: no authentication provided"),
				StatusCode: http.StatusUnauthorized,
			}
		}

		return authFunc(r.Context(), auth)
	}
}

/***********
 Basic Auth
***********/

func NewBasicAuthFunc(userPass map[string]string) AuthFunc {
	return basicAuth(userPass).authenticate
}

type basicAuth map[string]string

const AUTH_USERNAME_KEY = "request-username"

// basicAuth.authenticate meets the AuthFunc type
func (b basicAuth) authenticate(ctx context.Context, auth string) *Response {
	errResp := &Response{
		Err:        errors.New("unauthorized: invalid authentication provided"),
		StatusCode: http.StatusUnauthorized,
	}

	// parse the Authorization header
	u, p, ok := parseBasicAuth(auth)
	if !ok {
		return errResp
	}

	// get the password
	pass, ok := b[u]
	if !ok {
		return errResp
	}

	// compare the password
	if pass != p {
		return errResp
	}

	// add username to the context
	return &Response{
		Context: context.WithValue(ctx, AUTH_USERNAME_KEY, u),
	}
}

const basicPrefix = "Basic "

// parseBasicAuth parses an HTTP Basic Authentication string.
// taken from net/http/request.go
func parseBasicAuth(auth string) (username, password string, ok bool) {
	if !strings.HasPrefix(auth, basicPrefix) {
		return
	}
	c, err := base64.StdEncoding.DecodeString(auth[len(basicPrefix):])
	if err != nil {
		return
	}
	cs := string(c)
	s := strings.IndexByte(cs, ':')
	if s < 0 {
		return
	}
	return cs[:s], cs[s+1:], true
}

/****
 JWT
****/

type jwtAuth struct {
	secret string
}

func NewJWTAuthFunc(secret string) AuthFunc {
	j := &jwtAuth{secret: secret}
	return j.authenticate
}

const bearerPrefix = "Bearer "

func (j *jwtAuth) authenticate(ctx context.Context, auth string) *Response {
	// Remove 'Bearer' prefix
	if !strings.HasPrefix(auth, bearerPrefix) && !strings.HasPrefix(auth, strings.ToLower(bearerPrefix)) {
		return &Response{
			Err:        errors.New("unauthorized: invalid authentication provided"),
			StatusCode: http.StatusUnauthorized,
		}
	}

	token := auth[len(bearerPrefix):]

	_, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method")
		}
		return []byte(j.secret), nil
	})
	if err != nil {
		return &Response{
			Err:        err,
			StatusCode: http.StatusUnauthorized,
		}
	}

	return &Response{
		Context: context.WithValue(ctx, CONTEXT_JWT, token),
	}
}
