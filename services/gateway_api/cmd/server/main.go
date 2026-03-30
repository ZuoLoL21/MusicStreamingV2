package main

import (
	"gateway_api/internal/app"
	"gateway_api/internal/di"

	libsdi "libs/di"
	libsserver "libs/server"
)

func main() {
	logger := libsdi.InitLogger("gateway_api")
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
