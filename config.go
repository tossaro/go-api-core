package core

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
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

func NewConfig(f string) (Config, error) {
	cfg := Config{}
	var errB strings.Builder

	err := godotenv.Load(f)
	if err != nil {
		errB.WriteString("loading .env file")
	}

	aName, ok := os.LookupEnv("APP_NAME")
	if !ok {
		errB.WriteString("env APP_NAME not provided")
	}
	cfg.App.Name = aName

	aVersion, ok := os.LookupEnv("APP_VERSION")
	if !ok {
		errB.WriteString("env APP_VERSION not provided")
	}
	cfg.App.Version = aVersion

	hMode, ok := os.LookupEnv("HTTP_MODE")
	if !ok {
		errB.WriteString("env HTTP_MODE not provided")
	}
	cfg.HTTP.Mode = hMode

	hPort, ok := os.LookupEnv("HTTP_PORT")
	if !ok {
		errB.WriteString("env HTTP_PORT not provided")
	}
	cfg.HTTP.Port = hPort

	gPort, ok := os.LookupEnv("GRPC_PORT")
	if !ok {
		errB.WriteString("env GRPC_PORT not provided")
	}
	cfg.GRPC.Port = gPort

	logLevel, ok := os.LookupEnv("LOG_LEVEL")
	if !ok {
		errB.WriteString("env LOG_LEVEL not provided")
	}
	cfg.Log.Level = logLevel

	rUrl, ok := os.LookupEnv("REDIS_URL")
	if !ok {
		errB.WriteString("env REDIS_URL not provided")
	}
	cfg.Redis.Url = rUrl

	rPass, ok := os.LookupEnv("REDIS_PASSWORD")
	if !ok {
		errB.WriteString("env REDIS_PASSWORD not provided")
	}
	cfg.Redis.Password = rPass

	pUrl, ok := os.LookupEnv("POSTGRE_URL")
	if !ok {
		errB.WriteString("env POSTGRE_URL not provided")
	}
	cfg.Postgre.Url = pUrl

	pPoolMax, ok := os.LookupEnv("POSTGRE_POOL_MAX")
	if !ok {
		errB.WriteString("env POSTGRE_POOL_MAX not provided")
	}
	tPMIn, err := strconv.Atoi(pPoolMax)
	if err != nil {
		errB.WriteString(fmt.Sprintf("convert POSTGRE_POOL_MAX failed: %v", err))
	}
	cfg.Postgre.PoolMax = tPMIn

	tSID, ok := os.LookupEnv("TWILIO_SID")
	if !ok {
		errB.WriteString("env TWILIO_SID not provided")
	}
	cfg.Twilio.SID = tSID

	twToken, ok := os.LookupEnv("TWILIO_TOKEN")
	if !ok {
		errB.WriteString("env TWILIO_TOKEN not provided")
	}
	cfg.Twilio.SID = twToken

	twServiceSID, ok := os.LookupEnv("TWILIO_SERVICE_SID")
	if !ok {
		errB.WriteString("env TWILIO_SERVICE_SID not provided")
	}
	cfg.Twilio.SID = twServiceSID

	tAccess, ok := os.LookupEnv("TOKEN_ACCESS")
	if !ok {
		errB.WriteString("env TOKEN_ACCESS not provided")
	}
	tAcIn, err := strconv.Atoi(tAccess)
	if err != nil {
		errB.WriteString(fmt.Sprintf("convert TOKEN_ACCESS failed: %v", err))
	}
	cfg.TOKEN.Access = tAcIn

	tRefresh, ok := os.LookupEnv("TOKEN_REFRESH")
	if !ok {
		errB.WriteString("env TOKEN_REFRESH not provided")
	}
	tRefIn, err := strconv.Atoi(tRefresh)
	if err != nil {
		errB.WriteString(fmt.Sprintf("convert TOKEN_REFRESH failed: %v", err))
	}
	cfg.TOKEN.Refresh = tRefIn

	return cfg, fmt.Errorf(errB.String())
}
