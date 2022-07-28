package core

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/tossaro/go-api-core/logger"
)

type (
	Config struct {
		App struct {
			Name    string
			Version string
		}

		Services []Service

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

		Redis struct {
			Url      string
			Password string
		}

		Postgre struct {
			Url     string
			PoolMax int
		}

		Twilio struct {
			SID        string
			Token      string
			ServiceSID string
		}

		TOKEN struct {
			Access  int
			Refresh int
		}
	}

	Service struct {
		Url  string
		Name string
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
		log.Fatal("env APP_VERSION not provided\n")
	}
	cfg.App.Version = aVersion

	hMode, ok := os.LookupEnv("HTTP_MODE")
	if !ok {
		log.Fatal("env HTTP_MODE not provided\n")
	}
	cfg.HTTP.Mode = hMode

	hPort, ok := os.LookupEnv("HTTP_PORT")
	if !ok {
		log.Fatal("env HTTP_PORT not provided\n")
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
		l.Error("env GRPC_PORT not provided\n")
	}
	cfg.GRPC.Port = gPort

	rUrl, ok := os.LookupEnv("REDIS_URL")
	if !ok {
		l.Error("env REDIS_URL not provided\n")
	}
	cfg.Redis.Url = rUrl

	rPass, ok := os.LookupEnv("REDIS_PASSWORD")
	if !ok {
		l.Error("env REDIS_PASSWORD not provided\n")
	}
	cfg.Redis.Password = rPass

	pUrl, ok := os.LookupEnv("POSTGRE_URL")
	if !ok {
		l.Error("env POSTGRE_URL not provided\n")
	}
	cfg.Postgre.Url = pUrl

	pPoolMax, ok := os.LookupEnv("POSTGRE_POOL_MAX")
	if !ok {
		l.Error("env POSTGRE_POOL_MAX not provided\n")
	}
	tPMIn, err := strconv.Atoi(pPoolMax)
	if err != nil {
		l.Error(fmt.Sprintf("convert POSTGRE_POOL_MAX failed: %v\n", err))
	}
	cfg.Postgre.PoolMax = tPMIn

	tSID, ok := os.LookupEnv("TWILIO_SID")
	if !ok {
		l.Error("env TWILIO_SID not provided\n")
	}
	cfg.Twilio.SID = tSID

	twToken, ok := os.LookupEnv("TWILIO_TOKEN")
	if !ok {
		l.Error("env TWILIO_TOKEN not provided\n")
	}
	cfg.Twilio.Token = twToken

	twServiceSID, ok := os.LookupEnv("TWILIO_SERVICE_SID")
	if !ok {
		l.Error("env TWILIO_SERVICE_SID not provided\n")
	}
	cfg.Twilio.ServiceSID = twServiceSID

	tAccess, ok := os.LookupEnv("TOKEN_ACCESS")
	if !ok {
		l.Error("env TOKEN_ACCESS not provided\n")
	}
	tAcIn, err := strconv.Atoi(tAccess)
	if err != nil {
		l.Error(fmt.Sprintf("convert TOKEN_ACCESS failed: %v", err))
	}
	cfg.TOKEN.Access = tAcIn

	tRefresh, ok := os.LookupEnv("TOKEN_REFRESH")
	if !ok {
		l.Error("env TOKEN_REFRESH not provided\n")
	}
	tRefIn, err := strconv.Atoi(tRefresh)
	if err != nil {
		l.Error(fmt.Sprintf("convert TOKEN_REFRESH failed: %v", err))
	}
	cfg.TOKEN.Refresh = tRefIn

	sAuth, ok := os.LookupEnv("SERVICE_AUTH_URL")
	if !ok {
		l.Error("env SERVICE_AUTH_URL not provided\n")
	}
	cfg.Services = append(cfg.Services, Service{"Auth", sAuth})

	return cfg, l
}
