package core

import (
	"fmt"
	l "log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/tossaro/go-api-core/gin"
	"github.com/tossaro/go-api-core/httpserver"
	j "github.com/tossaro/go-api-core/jwt"
)

type (
	Options struct {
		EnvPath        string
		AuthType       string
		PrivateKeyPath *string
		PublicKeyPath  *string
		I18n           *i18n.Bundle
		Captcha        *bool
		Modules        []func(...interface{})
		ModuleParams   []interface{}
	}
)

func NewHttp(o Options) {
	if o.EnvPath == "" {
		l.Fatal("gin - EnvPath option not provided")
	}
	if o.AuthType == "" {
		l.Fatal("gin - AuthType option not provided")
	}
	if o.I18n == nil {
		l.Fatal("gin - I18n option not provided")
	}

	cfg, log := NewConfig(o.EnvPath)

	gOpt := gin.Options{
		I18n:     o.I18n,
		Mode:     cfg.HTTP.Mode,
		Version:  cfg.App.Version,
		BaseUrl:  cfg.App.Name,
		Log:      log,
		AuthType: o.AuthType,
		Captcha:  o.Captcha,
	}

	if o.AuthType == gin.AuthTypeGrpc {
		if len(cfg.Services) == 0 {
			log.Fatal("core - auth type grpc require auth service url")
		}
		gOpt.AuthService = &cfg.Services[0].Url
	} else if o.AuthType == gin.AuthTypeJwt {
		if o.PrivateKeyPath == nil || o.PublicKeyPath == nil {
			log.Fatal("core - auth type jwt require private and public key")
		}

		tAccess, ok := os.LookupEnv("TOKEN_ACCESS")
		if !ok {
			log.Error("env TOKEN_ACCESS not provided")
		}
		tAcIn, err := strconv.Atoi(tAccess)
		if err != nil {
			log.Error(fmt.Sprintf("convert TOKEN_ACCESS failed: %v", err))
		}

		tRefresh, ok := os.LookupEnv("TOKEN_REFRESH")
		if !ok {
			log.Error("env TOKEN_REFRESH not provided")
		}
		tRefIn, err := strconv.Atoi(tRefresh)
		if err != nil {
			log.Error(fmt.Sprintf("convert TOKEN_REFRESH failed: %v", err))
		}

		jwt := j.NewRSA(&j.Options{
			AccessTokenLifetime:  tAcIn,
			RefreshTokenLifetime: tRefIn,
			PrivateKeyPath:       *o.PrivateKeyPath,
			PublicKeyPath:        *o.PublicKeyPath,
		})
		gOpt.Jwt = jwt
	}

	g := gin.New(&gOpt)
	params := append(make([]interface{}, 0), g)
	for _, param := range o.ModuleParams {
		params = append(params, param)
	}
	for _, module := range o.Modules {
		module(params)
	}

	httpServer := httpserver.New(g.Gin, &httpserver.Options{
		Port: &cfg.HTTP.Port,
	})

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	var err error
	select {
	case s := <-interrupt:
		log.Info("core - signal: " + s.String())
	case err = <-httpServer.Notify():
		log.Error("core - notify http error: %s", err)
	}

	err = httpServer.Shutdown()
	if err != nil {
		log.Error("core - shutdown http error: %s", err)
	}
}
