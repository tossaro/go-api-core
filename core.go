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
	"github.com/tossaro/go-api-core/logger"
)

type (
	Options struct {
		Config         Config
		Log            logger.Interface
		AuthType       string
		AuthUrl        *string
		PrivateKeyPath *string
		PublicKeyPath  *string
		I18n           *i18n.Bundle
		Captcha        *bool
		Modules        []func([]interface{})
		ModuleParams   []interface{}
	}
)

func NewHttp(o Options) {
	if o.Config.App.Name == "" {
		l.Fatal("gin - Config option not provided")
	}
	if o.Log == nil {
		l.Fatal("gin - Log option not provided")
	}
	if o.AuthType == "" {
		l.Fatal("gin - AuthType option not provided")
	}
	if o.I18n == nil {
		l.Fatal("gin - I18n option not provided")
	}

	gOpt := gin.Options{
		I18n:     o.I18n,
		Mode:     o.Config.HTTP.Mode,
		Version:  o.Config.App.Version,
		BaseUrl:  o.Config.App.Name,
		Log:      o.Log,
		AuthType: o.AuthType,
		Captcha:  o.Captcha,
	}

	if o.AuthType == gin.AuthTypeGrpc {
		if o.AuthUrl == nil {
			o.Log.Fatal("core - auth type grpc require auth service url")
		}
		gOpt.AuthService = o.AuthUrl
	} else if o.AuthType == gin.AuthTypeJwt {
		if o.PrivateKeyPath == nil || o.PublicKeyPath == nil {
			o.Log.Fatal("core - auth type jwt require private and public key")
		}

		tAccess, ok := os.LookupEnv("TOKEN_ACCESS")
		if !ok {
			o.Log.Error("env TOKEN_ACCESS not provided")
		}
		tAcIn, err := strconv.Atoi(tAccess)
		if err != nil {
			o.Log.Error(fmt.Sprintf("convert TOKEN_ACCESS failed: %v", err))
		}

		tRefresh, ok := os.LookupEnv("TOKEN_REFRESH")
		if !ok {
			o.Log.Error("env TOKEN_REFRESH not provided")
		}
		tRefIn, err := strconv.Atoi(tRefresh)
		if err != nil {
			o.Log.Error(fmt.Sprintf("convert TOKEN_REFRESH failed: %v", err))
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
	params = append(params, o.ModuleParams...)
	l.Print(len(params))
	for _, module := range o.Modules {
		module(params)
	}

	httpServer := httpserver.New(g.Gin, &httpserver.Options{
		Port: &o.Config.HTTP.Port,
	})

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	var err error
	select {
	case s := <-interrupt:
		o.Log.Info("core - signal: " + s.String())
	case err = <-httpServer.Notify():
		o.Log.Error("core - notify http error: %s", err)
	}

	err = httpServer.Shutdown()
	if err != nil {
		o.Log.Error("core - shutdown http error: %s", err)
	}
}
