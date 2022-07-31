package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/nicksnyder/go-i18n/v2/i18n"

	core "github.com/tossaro/go-api-core"
	"github.com/tossaro/go-api-core/config"
	_ "github.com/tossaro/go-api-core/example/modular/docs"
	"github.com/tossaro/go-api-core/example/modular/internal/http"
	"github.com/tossaro/go-api-core/gin"
	"github.com/tossaro/go-api-core/logger"
	"github.com/tossaro/go-api-core/postgres"
	"golang.org/x/text/language"
)

// @title       API Core
// @description API Core
// @version     1.0.0
// @host        localhost:8888
// @BasePath    /go-api-core
func main() {
	cfg := config.New()
	log := logger.New(cfg)

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

	modules := []func([]interface{}){
		http.NewModule1V1,
	}

	params := append(
		make([]interface{}, 0),
		cfg,
		pg,
	)

	captcha := true
	privateKeyPath := "./key_private.pem"
	publicKeyPath := "./key_public.pem"
	core.NewHttp(core.Options{
		Config:         cfg,
		Log:            log,
		PrivateKeyPath: &privateKeyPath,
		PublicKeyPath:  &publicKeyPath,
		AuthType:       gin.AuthTypeJwt,
		I18n:           bI18n,
		Captcha:        &captcha,
		Modules:        modules,
		ModuleParams:   params,
	})
}
