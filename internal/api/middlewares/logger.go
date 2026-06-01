package middlewares

import (
	"sekai/pkg/ctxutils"

	"github.com/76Parker/golib/loglib"
	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
)

const requestIdHeader = "X-Request-ID"
const requestIdLogKey = "request_id"

func LoggerWithRequestID(logger loglib.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestID := ctx.GetHeader(requestIdHeader)
		if requestID == "" {
			requestID = xid.New().String()
		}
		logger := logger.With(requestIdLogKey, requestID)

		logCtx := ctxutils.SetRequestID(ctx.Request.Context(), requestID)
		logCtx = ctxutils.SetLogger(logCtx, logger)
		ctx.Request = ctx.Request.WithContext(logCtx)

		ctx.Header(requestIdHeader, requestID)
		ctx.Next()
	}
}
