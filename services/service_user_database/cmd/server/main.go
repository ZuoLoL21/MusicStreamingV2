package main

import (
	"backend/internal/app"
	"backend/internal/client"
	"backend/internal/di"
	sqlhandler "backend/sql/sqlc"
	"context"

	libsdi "libs/di"
	libsserver "libs/server"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func main() {
	logger := libsdi.InitLogger("service_user_database")
	defer func() {
		_ = logger.Sync()
	}()

	config := di.LoadConfig(logger)
	jwtHandler := libsdi.GetJWTHandler(logger, config, config.ApplicationName)
	returns := libsdi.NewReturnManager(logger)

	pool, err := pgxpool.New(context.Background(), config.DatabaseURL)
	if err != nil {
		logger.Fatal("failed to create connection pool", zap.Error(err))
	}
	defer pool.Close()
	db := sqlhandler.New(pool)

	fileStorage, err := client.NewMinIOFileStorageClient(
		config.MinIOEndpoint,
		config.MinIOAccessKey,
		config.MinIOSecretKey,
		config.MinIOBucketName,
		logger,
	)
	if err != nil {
		logger.Fatal("failed to create MinIO client", zap.Error(err))
	}

	clickhouseSync := client.NewClickHouseSync(logger, config, jwtHandler)

	application := app.New(logger, config, jwtHandler, returns, db, fileStorage, clickhouseSync)
	libsserver.RunHTTPServer(
		logger,
		":8080",
		application.Router(),
		libsserver.DefaultTimeouts(),
	)
}
