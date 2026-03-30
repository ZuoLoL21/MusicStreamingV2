package main

import (
	"gateway_recommendation/internal/app"
	"gateway_recommendation/internal/di"

	libsdi "libs/di"
	libsserver "libs/server"
)

func main() {
	logger := libsdi.InitLogger("gateway_recommendation")
	defer func() {
		_ = logger.Sync()
	}()

	config := di.LoadConfig(logger)
	jwtHandler := libsdi.GetJWTHandler(logger, config, config.ApplicationName)
	returnManager := libsdi.NewReturnManager(logger)

	application := app.New(config, logger, jwtHandler, returnManager)
	libsserver.RunHTTPServer(
		logger,
		":"+config.Port,
		application.Router(),
		libsserver.GatewayTimeouts(),
	)
}
