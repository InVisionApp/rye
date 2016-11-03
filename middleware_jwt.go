// This middleware provides JWT verification functionality
//
// You can use this middleware by specifying `rye.NewMiddlewareJWT(shared_secret)`
// when defining your routes.
//
// This middleware has no default version, it must be configured with a shared secret.
//
// Example use case:
//
// ```
//  routes.Handle("/some/route", a.Dependencies.MWHandler.Handle([]rye.Handler{
//     rye.NewMiddlewareJWT("this is a big secret"),
//     yourHandler,
// })).Methods("PUT", "OPTIONS")
// ```

package rye

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/dgrijalva/jwt-go"
)

type JwtVerify struct {
	secret string
	token  string
}

func NewMiddlewareJWT(secret string) func(rw http.ResponseWriter, req *http.Request) *Response {
	j := &JwtVerify{secret: secret}
	return j.handle
}

func (j *JwtVerify) handle(rw http.ResponseWriter, req *http.Request) *Response {
	tokenHeader := req.Header.Get("Authorization")

	// Remove 'Bearer' prefix
	p, _ := regexp.Compile(`(?i)bearer\s+`)
	j.token = p.ReplaceAllString(tokenHeader, "")

	_, err := jwt.Parse(j.token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method")
		}
		return []byte(j.secret), nil
	})

	if err != nil {
		return &Response{
			Err:        err,
			StatusCode: 401,
		}
	}

	return nil
}
