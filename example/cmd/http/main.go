package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-redis/redis/v8"
	core "github.com/tossaro/go-api-core"
	_ "github.com/tossaro/go-api-core/example/docs"
	"github.com/tossaro/go-api-core/gin"
	"github.com/tossaro/go-api-core/httpserver"
	j "github.com/tossaro/go-api-core/jwt"
	"github.com/tossaro/go-api-core/logger"
	"github.com/tossaro/go-api-core/postgres"
	"github.com/tossaro/go-api-core/twilio"
)

// @title       API Core
// @description API Core
// @version     1.0.0
// @host        localhost:8080
// @BasePath    /go-api-core
func main() {
	cfg, err := core.NewConfig("./../../.env")
	if err != nil {
		log.Printf("Config error: %s", err)
	}

	l := logger.New(cfg.Log.Level)

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Url,
		Password: cfg.Redis.Password,
		DB:       0,
	})
	defer rdb.Close()
	l.Info("app - redis initialized")

	_, err = postgres.New(&postgres.Options{
		Url: cfg.Postgre.Url,
	})
	if err != nil {
		l.Error("app - postgres.New: %w", err)
	}
	l.Info("app - postgres initialized")

	_, err = twilio.New(cfg.Twilio.SID, cfg.Twilio.Token, cfg.Twilio.ServiceSID)
	if err != nil {
		l.Error("app - twilio.New: %w", err)
	}
	l.Info("app - twilio initialized")

	jwt, err := j.NewRSA(&j.Options{
		AccessTokenLifetime:  cfg.TOKEN.Access,
		RefreshTokenLifetime: cfg.TOKEN.Refresh,
		PrivateKeyPath:       "./../../key_private.pem",
		PublicKeyPath:        "./../../key_public.pem",
	})
	if err != nil {
		l.Error("JWT error: %s", err)
	}

	cap := true
	// grpcUrl := ":" + cfg.GRPC.Port
	g := gin.New(&gin.Options{
		Mode:    cfg.HTTP.Mode,
		Version: cfg.App.Version,
		BaseUrl: cfg.App.Name,
		Logger:  l,
		// if session from redis enable redis & jwt
		Redis: rdb,
		Jwt:   jwt,
		// if session from another grpc service
		// AuthService:  &grpcUrl,
		AccessToken:  cfg.TOKEN.Access,
		RefreshToken: cfg.TOKEN.Refresh,
		Captcha:      &cap,
	})

	httpServer := httpserver.New(g.Gin, &httpserver.Options{
		Port: &cfg.HTTP.Port,
	})

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		l.Info("app - signal: " + s.String())
	case err = <-httpServer.Notify():
		l.Error(fmt.Errorf("app - httpServer.Notify: %w", err))
	}

	err = httpServer.Shutdown()
	if err != nil {
		l.Error(fmt.Errorf("app - httpServer.Shutdown: %w", err))
	}
}
