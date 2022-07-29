package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	core "github.com/tossaro/go-api-core"
	_ "github.com/tossaro/go-api-core/example/manual/docs"
	"github.com/tossaro/go-api-core/example/manual/internal/module1"
	"github.com/tossaro/go-api-core/gin"
	"github.com/tossaro/go-api-core/httpserver"
	j "github.com/tossaro/go-api-core/jwt"
	"github.com/tossaro/go-api-core/postgres"
	"github.com/tossaro/go-api-core/twilio"
	"golang.org/x/text/language"
)

// @title       API Core
// @description API Core
// @version     1.0.0
// @host        localhost:8888
// @BasePath    /go-api-core
func main() {
	var err error
	cfg, log := core.NewConfig("./.env")

	twSID, ok := os.LookupEnv("TWILIO_SID")
	if !ok {
		log.Fatal("env TWILIO_SID not provided")
	}

	twToken, ok := os.LookupEnv("TWILIO_TOKEN")
	if !ok {
		log.Fatal("env TWILIO_TOKEN not provided")
	}

	twServiceSID, ok := os.LookupEnv("TWILIO_SERVICE_SID")
	if !ok {
		log.Fatal("env TWILIO_SERVICE_SID not provided")
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

	bI18n := i18n.NewBundle(language.English)
	bI18n.RegisterUnmarshalFunc("json", json.Unmarshal)
	bI18n.MustLoadMessageFile("./i18n/en.json")
	bI18n.MustLoadMessageFile("./i18n/id.json")

	pg := postgres.New(&postgres.Options{
		Url: cfg.Postgre.Url,
	})
	log.Info("app - postgres initialized")

	twl := twilio.New(&twilio.Options{
		SID:        twSID,
		Token:      twToken,
		ServiceSID: twServiceSID,
	})
	log.Info("app - twilio initialized")

	jwt := j.NewRSA(&j.Options{
		AccessTokenLifetime:  tAcIn,
		RefreshTokenLifetime: tRefIn,
		PrivateKeyPath:       "./key_private.pem",
		PublicKeyPath:        "./key_public.pem",
	})

	captcha := true
	g := gin.New(&gin.Options{
		I18n:     bI18n,
		Mode:     cfg.HTTP.Mode,
		Version:  cfg.App.Version,
		BaseUrl:  cfg.App.Name,
		Log:      log,
		AuthType: gin.AuthTypeJwt,
		// if auth type jwt
		Jwt: jwt,
		// if auth type grpc
		// AuthService:  &cfg.Services[0].Url,
		Captcha: &captcha,
	})

	module1.NewHttpV1(g, pg, twl)

	httpServer := httpserver.New(g.Gin, &httpserver.Options{
		Port: &cfg.HTTP.Port,
	})

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		log.Info("app - signal: " + s.String())
	case err = <-httpServer.Notify():
		log.Error("app - httpServer.Notify: %s", err)
	}

	err = httpServer.Shutdown()
	if err != nil {
		log.Error("app - httpServer.Shutdown: %s", err)
	}
}
