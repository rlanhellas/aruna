package httpbridge

import (
	"context"
	"github.com/gin-gonic/gin"
)

type RouteGroupHttp struct {
	Path          string
	Authenticated bool
	Routes        []*RouteHttp
}

// RouteHttp configure a http route
type RouteHttp struct {
	Path                  string
	Handler               func(ctx context.Context, in any, ginctx *gin.Context) *HandlerHttpResponse
	HandlerInputGenerator func() any //func to generate new input instance to be bound in HTTP request
	Method                string
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
