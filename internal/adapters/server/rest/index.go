package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rendau/dop/adapters/logger"
	dopHttps "github.com/rendau/dop/adapters/server/https"
	swagFiles "github.com/swaggo/files"
	ginSwag "github.com/swaggo/gin-swagger"

	"github.com/rendau/kazan/internal/domain/core"
)

type St struct {
	lg   logger.Lite
	core *core.St
}

func GetHandler(lg logger.Lite, core *core.St, withCors bool) http.Handler {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()

	// middlewares

	r.Use(dopHttps.MwRecovery(lg, nil))
	if withCors {
		r.Use(dopHttps.MwCors())
	}

	// handlers

	// doc
	r.GET("/doc/*any", ginSwag.WrapHandler(swagFiles.Handler, func(c *ginSwag.Config) {
		c.DefaultModelsExpandDepth = 0
		c.DocExpansion = "none"
	}))

	s := &St{lg: lg, core: core}

	// healthcheck
	r.GET("/healthcheck", func(c *gin.Context) { c.Status(http.StatusOK) })

	// static
	r.POST("/static", s.hStaticSave)
	r.GET("/static/*any", s.hStaticGet)

	// kvs
	r.POST("/kvs/:key", s.hKvsSet)
	r.GET("/kvs/:key", s.hKvsGet)
	r.DELETE("/kvs/:key", s.hKvsRemove)

	// clean
	r.GET("/clean", s.hClean)

	return r
}
