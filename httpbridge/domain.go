package httpbridge

import "context"

// RouteHttp configure a http route
type RouteHttp struct {
	Path         string
	Handler      func(in any, ctx context.Context) *HandlerHttpResponse
	HandlerInput any //data sent from http request will be unmarshal to this input
	Method       string
}

// BaseHttpResponse json which will be rendered in response for the user
type BaseHttpResponse struct {
	ErrorMessage string `json:"error_message"`
	Data         any    `json:"data"`
}

// HandlerHttpResponse handler response to be transformed in BaseHttpResponse later on
type HandlerHttpResponse struct {
	Error      error
	Data       any
	StatusCode int
}
