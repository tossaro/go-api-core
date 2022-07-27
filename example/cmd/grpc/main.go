package main

import (
	"fmt"
	"log"
	"net"

	core "github.com/tossaro/go-api-core"
	"github.com/tossaro/go-api-core/grpc/auth"
	"github.com/tossaro/go-api-core/logger"
	"google.golang.org/grpc"
)

type (
	authServer struct {
		auth.UnimplementedAuthServiceServer
	}
)

func main() {
	cfg, err := core.NewConfig("./../../.env")
	if err != nil {
		log.Printf("Config error: %s", err)
	}

	l := logger.New(cfg.Log.Level)

	conn, err := net.Listen("tcp", ":"+cfg.GRPC.Port)
	if err != nil {
		l.Error(fmt.Errorf("app - tcp connection error: %w", err))
	}

	s := grpc.NewServer()
	auth.RegisterAuthServiceServer(s, &authServer{})
	l.Info("app - grpc listening at %v", conn.Addr())
	if err := s.Serve(conn); err != nil {
		l.Error(fmt.Errorf("app - failed serve grpc: %w", err))
	}
}
