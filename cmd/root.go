package cmd

import (
	"os"
	"time"

	dopLoggerZap "github.com/rendau/dop/adapters/logger/zap"
	dopServerHttps "github.com/rendau/dop/adapters/server/https"
	"github.com/rendau/dop/dopTools"

	"github.com/rendau/kazan/docs"
	"github.com/rendau/kazan/internal/adapters/server/rest"
	"github.com/rendau/kazan/internal/domain/core"
)

func Execute() {
	// var err error

	app := struct {
		lg         *dopLoggerZap.St
		core       *core.St
		restApi    *rest.St
		restApiSrv *dopServerHttps.St
	}{}

	confLoad()

	app.lg = dopLoggerZap.New(conf.LogLevel, conf.Debug)

	app.core = core.New(
		app.lg,
		conf.ImgMaxWidth,
		conf.ImgMaxHeight,
		false,
	)

	docs.SwaggerInfo.Host = conf.SwagHost
	docs.SwaggerInfo.BasePath = conf.SwagBasePath
	docs.SwaggerInfo.Schemes = []string{conf.SwagSchema}
	docs.SwaggerInfo.Title = "FS service"

	// START

	app.lg.Infow("Starting")

	app.core.Start()

	app.restApiSrv = dopServerHttps.Start(
		conf.HttpListen,
		rest.GetHandler(
			app.lg,
			app.core,
			conf.HttpCors,
		),
		app.lg,
	)

	var exitCode int

	select {
	case <-dopTools.StopSignal():
	case <-app.restApiSrv.Wait():
		exitCode = 1
	}

	// STOP

	app.lg.Infow("Shutting down...")

	if !app.restApiSrv.Shutdown(20 * time.Second) {
		exitCode = 1
	}

	app.lg.Infow("Wait routines...")

	app.core.StopAndWaitJobs()

	app.lg.Infow("Exit")

	os.Exit(exitCode)
}
