package core

import (
	"fmt"
	"os"
	"strconv"

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

func NewConfig() (Config, error) {
	cfg := Config{}

	err := godotenv.Load()
	if err != nil {
		return cfg, fmt.Errorf("loading .env file")
	}

	aName, ok := os.LookupEnv("APP_NAME")
	if !ok {
		return cfg, fmt.Errorf("env APP_NAME not provided")
	}
	cfg.App.Name = aName

	aVersion, ok := os.LookupEnv("APP_VERSION")
	if !ok {
		return cfg, fmt.Errorf("env APP_VERSION not provided")
	}
	cfg.App.Version = aVersion

	hMode, ok := os.LookupEnv("HTTP_MODE")
	if !ok {
		return cfg, fmt.Errorf("env HTTP_MODE not provided")
	}
	cfg.HTTP.Mode = hMode

	hPort, ok := os.LookupEnv("HTTP_PORT")
	if !ok {
		return cfg, fmt.Errorf("env HTTP_PORT not provided")
	}
	cfg.HTTP.Port = hPort

	logLevel, ok := os.LookupEnv("LOG_LEVEL")
	if !ok {
		return cfg, fmt.Errorf("env LOG_LEVEL not provided")
	}
	cfg.Log.Level = logLevel

	rUrl, ok := os.LookupEnv("REDIS_URL")
	if !ok {
		return cfg, fmt.Errorf("env REDIS_URL not provided")
	}
	cfg.Redis.Url = rUrl

	rPass, ok := os.LookupEnv("REDIS_PASSWORD")
	if !ok {
		return cfg, fmt.Errorf("env REDIS_PASSWORD not provided")
	}
	cfg.Redis.Password = rPass

	pUrl, ok := os.LookupEnv("POSTGRE_URL")
	if !ok {
		return cfg, fmt.Errorf("env POSTGRE_URL not provided")
	}
	cfg.Postgre.Url = pUrl

	pPoolMax, ok := os.LookupEnv("POSTGRE_POOL_MAX")
	if !ok {
		return cfg, fmt.Errorf("env POSTGRE_POOL_MAX not provided")
	}
	tPMIn, err := strconv.Atoi(pPoolMax)
	if err != nil {
		return cfg, fmt.Errorf("convert POSTGRE_POOL_MAX failed: %v", err)
	}
	cfg.Postgre.PoolMax = tPMIn

	tSID, ok := os.LookupEnv("TWILIO_SID")
	if !ok {
		return cfg, fmt.Errorf("env TWILIO_SID not provided")
	}
	cfg.Twilio.SID = tSID

	twToken, ok := os.LookupEnv("TWILIO_TOKEN")
	if !ok {
		return cfg, fmt.Errorf("env TWILIO_TOKEN not provided")
	}
	cfg.Twilio.SID = twToken

	twServiceSID, ok := os.LookupEnv("TWILIO_SERVICE_SID")
	if !ok {
		return cfg, fmt.Errorf("env TWILIO_SERVICE_SID not provided")
	}
	cfg.Twilio.SID = twServiceSID

	tAccess, ok := os.LookupEnv("TOKEN_ACCESS")
	if !ok {
		return cfg, fmt.Errorf("env TOKEN_ACCESS not provided")
	}
	tAcIn, err := strconv.Atoi(tAccess)
	if err != nil {
		return cfg, fmt.Errorf("convert TOKEN_ACCESS failed: %v", err)
	}
	cfg.TOKEN.Access = tAcIn

	tRefresh, ok := os.LookupEnv("TOKEN_REFRESH")
	if !ok {
		return cfg, fmt.Errorf("env TOKEN_REFRESH not provided")
	}
	tRefIn, err := strconv.Atoi(tRefresh)
	if err != nil {
		return cfg, fmt.Errorf("convert TOKEN_REFRESH failed: %v", err)
	}
	cfg.TOKEN.Refresh = tRefIn

	return cfg, nil
}
