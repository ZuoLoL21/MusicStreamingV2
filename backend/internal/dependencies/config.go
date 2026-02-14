package dependencies

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

type Config struct {
	Provider             string
	SubjectNormal        string
	JWTExpirationNormal  time.Duration
	SubjectRefresh       string
	JWTExpirationRefresh time.Duration
}

func LoadConfig(logger *zap.Logger) *Config {
	slogger := logger.With(
		zap.String("lifespan", "init"),
	).Sugar()

	err := godotenv.Load()
	if err != nil {
		slogger.Errorf("Error loading .env file: %v", err)
	}

	normalTime, err := strconv.Atoi(os.Getenv("TIME_IN_M_NORMAL"))
	if err != nil {
		slogger.Errorf("Error parsing TIME_IN_M_NORMAL: %v", err)
	}

	refreshTime, err := strconv.Atoi(os.Getenv("TIME_IN_D_REFRESH"))
	if err != nil {
		slogger.Errorf("Error parsing TIME_IN_D_REFRESH: %v", err)
	}

	return &Config{
		Provider:             os.Getenv("PROVIDER"),
		SubjectNormal:        os.Getenv("SUBJECT_NORMAL"),
		JWTExpirationNormal:  time.Minute * time.Duration(normalTime),
		SubjectRefresh:       os.Getenv("SUBJECT_REFRESH"),
		JWTExpirationRefresh: time.Hour * 24 * time.Duration(refreshTime),
	}
}
