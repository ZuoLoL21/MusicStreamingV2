package main

import (
	"event_ingestion/internal/app"
	"event_ingestion/internal/di"

	libsdi "libs/di"
	libsserver "libs/server"

	"go.uber.org/zap"
)

func main() {
	logger := libsdi.InitLogger("service_event_ingestion")
	defer func() {
		_ = logger.Sync()
	}()

	config := di.LoadConfig(logger)
	jwtHandler := libsdi.GetJWTHandler(logger, config, config.ApplicationName)
	returns := libsdi.NewReturnManager(logger)

	clickhouse, err := di.NewClickHouseClient(config, logger)
	if err != nil {
		logger.Fatal("failed to create ClickHouse client", zap.Error(err))
	}
	defer func(clickhouse *di.ClickHouseClient) {
		_ = clickhouse.Close()
	}(clickhouse)

	application := app.New(logger, config, jwtHandler, returns, clickhouse)
	libsserver.RunHTTPServer(
		logger,
		":"+config.Port,
		application.Router(),
		libsserver.DefaultTimeouts(),
	)
}
