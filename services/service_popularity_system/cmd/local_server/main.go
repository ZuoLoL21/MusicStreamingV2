package main

import (
	"popularity/internal/app"
	"popularity/internal/di"

	libsdi "libs/di"
	libsserver "libs/server"
)

func main() {
	logger := libsdi.InitLogger("service_popularity")
	defer func() {
		_ = logger.Sync()
	}()

	config := di.LoadConfig(logger)
	jwtHandler := libsdi.GetJWTHandler(logger, config, config.ApplicationName)
	returns := libsdi.NewReturnManager(logger)

	application := app.New(logger, config, jwtHandler, returns)
	libsserver.RunHTTPServer(
		logger,
		":"+config.Port,
		application.Router(),
		libsserver.DefaultTimeouts(),
	)
}
