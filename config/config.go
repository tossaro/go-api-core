package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

const (
	LogTypeStdOut = "stdout"
	LogTypeFile   = "file"

	_defaultEnvPath       = "./.env"
	_defaultLogFileName   = "core.log"
	_defaultLogMaxSize    = 100 //mb
	_defaultLogMaxAge     = 10  //day
	_defaultLogMaxBackups = 10  //file
	_defaultLogCompress   = false
)

type Config struct {
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
		Type       string
		Level      string
		FileName   string
		MaxSize    int
		MaxAge     int
		MaxBackups int
		Compress   bool
	}
}

func New(f ...string) Config {
	p := _defaultEnvPath
	if len(f) > 0 {
		p = f[0]
	}

	cfg := Config{}
	err := godotenv.Load(p)
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

	gPort, ok := os.LookupEnv("GRPC_PORT")
	if !ok {
		log.Fatal("env GRPC_PORT not provided")
	}
	cfg.GRPC.Port = gPort

	logLevel, ok := os.LookupEnv("LOG_LEVEL")
	if !ok {
		log.Fatal("env LOG_LEVEL not provided")
	}
	cfg.Log.Level = logLevel

	logType, ok := os.LookupEnv("LOG_TYPE")
	if !ok {
		log.Fatal("env LOG_TYPE not provided")
	}
	cfg.Log.Type = logType

	if logType == LogTypeFile {
		cfg.Log.FileName = _defaultLogFileName
		logFileName, ok := os.LookupEnv("LOG_FILE_NAME")
		if !ok {
			log.Print("env LOG_FILE_NAME not provided, using default = " + _defaultLogFileName)
		} else {
			cfg.Log.FileName = logFileName
		}

		cfg.Log.MaxSize = _defaultLogMaxSize
		logMaxSize, ok := os.LookupEnv("LOG_MAX_SIZE")
		if !ok {
			log.Print("env LOG_MAX_SIZE not provided, using default = " + strconv.Itoa(_defaultLogMaxSize))
		} else {
			logMaxSizeIn, err := strconv.Atoi(logMaxSize)
			if err != nil {
				log.Fatal(fmt.Sprintf("convert LOG_MAX_SIZE failed: %v", err))
			}
			cfg.Log.MaxSize = logMaxSizeIn
		}

		cfg.Log.MaxAge = _defaultLogMaxAge
		logMaxAge, ok := os.LookupEnv("LOG_MAX_AGE")
		if !ok {
			log.Print("env LOG_MAX_AGE not provided, using default = " + strconv.Itoa(_defaultLogMaxAge))
		} else {
			logMaxAgeIn, err := strconv.Atoi(logMaxAge)
			if err != nil {
				log.Fatal(fmt.Sprintf("convert LOG_MAX_AGE failed: %v", err))
			}
			cfg.Log.MaxAge = logMaxAgeIn
		}

		cfg.Log.MaxBackups = _defaultLogMaxBackups
		logMaxBackups, ok := os.LookupEnv("LOG_MAX_BACKUPS")
		if !ok {
			log.Print("env LOG_MAX_BACKUPS not provided, using default = " + strconv.Itoa(_defaultLogMaxBackups))
		} else {
			logMaxBackupsIn, err := strconv.Atoi(logMaxBackups)
			if err != nil {
				log.Fatal(fmt.Sprintf("convert LOG_MAX_BACKUPS failed: %v", err))
			}
			cfg.Log.MaxBackups = logMaxBackupsIn
		}

		cfg.Log.Compress = _defaultLogCompress
		logCompress, ok := os.LookupEnv("LOG_COMPRESS")
		if !ok {
			log.Print("env LOG_COMPRESS not provided, using default = " + strconv.FormatBool(_defaultLogCompress))
		} else {
			logCompressBool, err := strconv.ParseBool(logCompress)
			if err != nil {
				log.Fatal(fmt.Sprintf("convert LOG_COMPRESS failed: %v", err))
			}
			cfg.Log.Compress = logCompressBool
		}
	}

	return cfg
}
