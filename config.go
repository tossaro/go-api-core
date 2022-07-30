package core

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/tossaro/go-api-core/logger"
)

type (
	Config struct {
		App struct {
			Name    string
			Version string
		}

		HTTP struct {
			Mode string
			Port string
		}

		GRPC struct {
			Port string
		}

		Log struct {
			Level string
		}
	}
)

func NewConfig(f string) (Config, logger.Interface) {
	cfg := Config{}

	err := godotenv.Load(f)
	if err != nil {
		log.Fatal("error loading config file")
	}

	aName, ok := os.LookupEnv("APP_NAME")
	if !ok {
		log.Fatal("env APP_NAME not provided")
	}
	cfg.App.Name = aName

	aVersion, ok := os.LookupEnv("APP_VERSION")
	if !ok {
		log.Fatal("env APP_VERSION not provided")
	}
	cfg.App.Version = aVersion

	hMode, ok := os.LookupEnv("HTTP_MODE")
	if !ok {
		log.Fatal("env HTTP_MODE not provided")
	}
	cfg.HTTP.Mode = hMode

	hPort, ok := os.LookupEnv("HTTP_PORT")
	if !ok {
		log.Fatal("env HTTP_PORT not provided")
	}
	cfg.HTTP.Port = hPort

	logLevel, ok := os.LookupEnv("LOG_LEVEL")
	if !ok {
		log.Fatal("env LOG_LEVEL not provided")
	}
	cfg.Log.Level = logLevel

	l := logger.New(logLevel)

	gPort, ok := os.LookupEnv("GRPC_PORT")
	if !ok {
		l.Error("env GRPC_PORT not provided")
	}
	cfg.GRPC.Port = gPort

	return cfg, l
}
