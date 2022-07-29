package main

import (
	"encoding/json"
	"log"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	core "github.com/tossaro/go-api-core"
	"github.com/tossaro/go-api-core/example/modular/internal/module1"
	"github.com/tossaro/go-api-core/gin"
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
	bI18n := i18n.NewBundle(language.English)
	bI18n.RegisterUnmarshalFunc("json", json.Unmarshal)
	bI18n.MustLoadMessageFile("./i18n/en.json")
	bI18n.MustLoadMessageFile("./i18n/id.json")

	pg := postgres.New(&postgres.Options{
		Url: "",
	})
	log.Println("app - postgres initialized")

	twl := twilio.New(&twilio.Options{
		SID:        "",
		Token:      "",
		ServiceSID: "",
	})
	log.Println("app - twilio initialized")

	captcha := true
	privateKeyPath := "./key_private.pem"
	publicKeyPath := "./key_public.pem"
	core.NewHttp(core.Options{
		EnvPath:        "./.env",
		PrivateKeyPath: &privateKeyPath,
		PublicKeyPath:  &publicKeyPath,
		AuthType:       gin.AuthTypeJwt,
		I18n:           bI18n,
		Captcha:        &captcha,
		Modules:        []func(...interface{}){module1.NewHttpV1},
		ModuleParams:   append(make([]interface{}, 0), twl, pg),
	})
}
