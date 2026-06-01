package api

import (
	"net/http"
	"sekai/internal/api/middlewares"

	"github.com/76Parker/golib/loglib"
	"github.com/MicahParks/keyfunc/v3"
	"github.com/gin-gonic/gin"
)

type RouterConfig struct {
	AuthEnabled bool
	AuthIssuer  string
}

func NewRouter(
	cfg RouterConfig,
	logger loglib.Logger,
	jwks keyfunc.Keyfunc,
	userRepository middlewares.UserRepository,
	scanHandler *ScanHandler,
) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middlewares.LoggerWithRequestID(logger))
	router.Use(middlewares.ErrorHandler())

	router.GET("/swagger", swaggerUI)
	router.GET("/swagger/", swaggerUI)
	router.GET("/swagger/index.html", swaggerUI)
	router.GET("/swagger/openapi.yaml", openAPIYAML)
	router.GET("/docs/OpenAPI.yaml", openAPIYAML)
	router.GET("/docs/openapi.yaml", openAPIYAML)

	router.GET("/healthz", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	protected := router.Group("/")
	if cfg.AuthEnabled {
		protected.Use(middlewares.Auth(jwks, cfg.AuthIssuer, userRepository))
	} else {
		protected.Use(middlewares.NoAuth(userRepository))
	}
	protected.POST("/scans", scanHandler.StartScan)
	protected.GET("/scans", scanHandler.ListScans)

	return router
}

func swaggerUI(ctx *gin.Context) {
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Sekai API Swagger</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
  <style>
    body { margin: 0; background: #fafafa; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.ui = SwaggerUIBundle({
      url: "/swagger/openapi.yaml",
      dom_id: "#swagger-ui",
      deepLinking: true,
      persistAuthorization: true
    });
  </script>
</body>
</html>`))
}

func openAPIYAML(ctx *gin.Context) {
	ctx.Header("Cache-Control", "no-store")
	ctx.File("docs/OpenAPI.yaml")
}
