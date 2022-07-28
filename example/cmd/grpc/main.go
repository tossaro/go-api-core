package main

import (
	"net"

	core "github.com/tossaro/go-api-core"
	pAuth "github.com/tossaro/go-api-core/auth/proto"
	"google.golang.org/grpc"
)

type (
	service1 struct {
		pAuth.UnimplementedAuthServiceV1Server
	}
)

func main() {
	cfg, log := core.NewConfig("./.env")

	conn, err := net.Listen("tcp", ":"+cfg.GRPC.Port)
	if err != nil {
		log.Fatal("app - tcp connection error: %s", err)
	}

	s := grpc.NewServer()
	pAuth.RegisterAuthServiceV1Server(s, &service1{})

	log.Info("app - grpc listening at %v", conn.Addr())
	if err := s.Serve(conn); err != nil {
		log.Error("app - failed serve grpc: %s", err)
	}
}
