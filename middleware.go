package eotel

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
)

func Middleware(name string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, span := otel.Tracer(globalCfg.ServiceName).
			Start(c.Request.Context(), fmt.Sprintf("%s %s", c.Request.Method, c.FullPath()))
		defer span.End()

		logger := New(ctx, name).
			WithField("method", c.Request.Method).
			WithField("path", c.Request.URL.Path).
			WithField("ip", c.ClientIP()).
			WithField("ua", c.Request.UserAgent())

		ctx = logger.Inject(ctx, logger)
		c.Request = c.Request.WithContext(ctx)

		defer logger.RecoverPanic(c)
		c.Next()
	}
}
