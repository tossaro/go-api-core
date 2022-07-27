package main

import (
	"fmt"
	"log"
	"net"

	core "github.com/tossaro/go-api-core"
	pAuth "github.com/tossaro/go-api-core/auth/proto"
	"github.com/tossaro/go-api-core/logger"
	"google.golang.org/grpc"
)

type (
	authServer struct {
		pAuth.UnimplementedAuthServiceV1Server
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
	pAuth.RegisterAuthServiceV1Server(s, &authServer{})

	l.Info("app - grpc listening at %v", conn.Addr())
	if err := s.Serve(conn); err != nil {
		l.Error(fmt.Errorf("app - failed serve grpc: %w", err))
	}
}
