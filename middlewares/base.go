package middlewares

// Response struct is utilized by middlewares as a way to share state;
// ie. a middleware can return a *middleware.Response as a way to indicate
// that further middleware execution should stop (without an error) or return a
// a hard error by setting `Err` + `StatusCode`.
type Response struct {
	Err           error
	StatusCode    int
	StopExecution bool
}

// Meet the Error interface
func (r *Response) Error() string {
	return r.Err.Error()
}
