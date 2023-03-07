package httpbridge

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rlanhellas/aruna/global"
	"github.com/rlanhellas/aruna/logger"
)

// HttpHandler handler for all http requests
func HttpHandler(ginctx *gin.Context, ctx context.Context, routeHttp *RouteHttp) {
	form, errForm := ginctx.MultipartForm()
	if errForm != nil {
		logger.Error(ctx, "error to collect files: %s", errForm.Error())
	}
	newCtx := context.WithValue(ctx, global.CorrelationID, ginctx.GetHeader(global.CorrelationID))
	if form != nil {
		newCtx = context.WithValue(newCtx, global.RequestForm, form)
	}

	logger.Debug(newCtx, "handling path %s, method %s", routeHttp.Path, routeHttp.Method)

	var in any
	if routeHttp.HandlerInputGenerator != nil {
		in = routeHttp.HandlerInputGenerator()
		err := ginctx.BindJSON(in)
		if err != nil {
			logger.Error(newCtx, "error trying to bindJson. %s", err.Error())
			ginctx.JSON(http.StatusBadRequest, BaseHttpResponse{
				ErrorMessage: fmt.Sprintf("can not bind your object to json. %s", err.Error()),
			})
			return
		}
	}

	handlerResponse := routeHttp.Handler(newCtx, in, ginctx)
	logger.Debug(newCtx, "handler response status code %d. error: %v", handlerResponse.StatusCode, handlerResponse.Error)
	baseHttpResponse := BaseHttpResponse{}
	baseHttpResponse.Data = handlerResponse.Data
	if handlerResponse.Error != nil {
		baseHttpResponse.ErrorMessage = handlerResponse.Error.Error()
	}

	ginctx.JSON(handlerResponse.StatusCode, baseHttpResponse)
}
