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

	tAccess, ok := os.LookupEnv("TOKEN_ACCESS")
	if !ok {
		log.Fatal("env TOKEN_ACCESS not provided")
	}
	tAcIn, err := strconv.Atoi(tAccess)
	if err != nil {
		log.Fatal(fmt.Sprintf("convert TOKEN_ACCESS failed: %v", err))
	}
	tRefresh, ok := os.LookupEnv("TOKEN_REFRESH")
	if !ok {
		log.Fatal("env TOKEN_REFRESH not provided")
	}
	tRefIn, err := strconv.Atoi(tRefresh)
	if err != nil {
		log.Fatal(fmt.Sprintf("convert TOKEN_REFRESH failed: %v", err))
	}
	pUrl, ok := os.LookupEnv("POSTGRE_URL")
	if !ok {
		log.Fatal("env POSTGRE_URL not provided")
	}
	pPoolMax, ok := os.LookupEnv("POSTGRE_POOL_MAX")
	if !ok {
		log.Fatal("env POSTGRE_POOL_MAX not provided")
	}
	pPoolMaxIn, err := strconv.Atoi(pPoolMax)
	if err != nil {
		log.Fatal(fmt.Sprintf("convert POSTGRE_POOL_MAX failed: %v", err))
	}

	bI18n := i18n.NewBundle(language.English)
	bI18n.RegisterUnmarshalFunc("json", json.Unmarshal)
	bI18n.MustLoadMessageFile("./i18n/en.json")
	bI18n.MustLoadMessageFile("./i18n/id.json")

	pg := postgres.New(&postgres.Options{
		Url:     pUrl,
		PoolMax: &pPoolMaxIn,
	})
	log.Info("app - postgres initialized")

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

	module1.NewHttpV1(g, pg)

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
