package middlewares

import (
	"errors"
	"net/http"
	"sekai/internal/entities/dto"
	"sekai/pkg/ctxutils"

	"github.com/gin-gonic/gin"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) == 0 {
			return
		}
		err := c.Errors.Last().Err

		requestID := ctxutils.RequestID(c.Request.Context())
		c.Header(requestIdHeader, requestID)

		if e, ok := errors.AsType[dto.APIError](err); ok {
			c.JSON(e.Status, e)
			if e.Status == http.StatusInternalServerError {
				c.Status(http.StatusInternalServerError)
				return
			}
			return
		}
		c.Status(http.StatusInternalServerError)
		c.Next()
	}
}
