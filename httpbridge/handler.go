package httpbridge

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/rlanhellas/aruna/global"
	"github.com/rlanhellas/aruna/logger"
	"net/http"
)

// HttpHandler handler for all http requests
func HttpHandler(ginctx *gin.Context, ctx context.Context, routeHttp *RouteHttp) {
	newCtx := context.WithValue(ctx, global.CorrelationID, ginctx.GetHeader(global.CorrelationID))

	logger.Debug(newCtx, "handling path %s, method %s", routeHttp.Path, routeHttp.Method)

	err := ginctx.BindJSON(routeHttp.HandlerInput)
	if err != nil {
		logger.Error(newCtx, "error trying to bindJson. %s", err.Error())
		ginctx.JSON(http.StatusBadRequest, BaseHttpResponse{
			ErrorMessage: "can not bind your object to json",
		})
		return
	}

	handlerResponse := routeHttp.Handler(routeHttp.HandlerInput, newCtx)
	logger.Debug(newCtx, "handler response status code %d. error: %v", handlerResponse.StatusCode, handlerResponse.Error)
	ginctx.JSON(handlerResponse.StatusCode, BaseHttpResponse{
		ErrorMessage: handlerResponse.Error.Error(),
		Data:         handlerResponse.Data,
	})
}
