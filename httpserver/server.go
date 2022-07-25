package httpserver

import (
	"context"
	"net/http"
	"time"
)

const (
	_defaultReadTimeout     = 5 * time.Second
	_defaultWriteTimeout    = 5 * time.Second
	_defaultAddr            = ":80"
	_defaultShutdownTimeout = 3 * time.Second
)

type (
	Options struct {
		Port            *string
		ReadTimeout     *time.Duration
		WriteTimeout    *time.Duration
		ShutDownTimeout *time.Duration
	}

	Server struct {
		server          *http.Server
		notify          chan error
		shutdownTimeout time.Duration
	}
)

func New(handler http.Handler, opts *Options) *Server {
	rT := _defaultReadTimeout
	if opts.ReadTimeout != nil {
		rT = *(opts.ReadTimeout)
	}
	wT := _defaultWriteTimeout
	if opts.WriteTimeout != nil {
		wT = *(opts.WriteTimeout)
	}
	a := _defaultAddr
	if opts.Port != nil {
		a = ":" + *(opts.Port)
	}

	httpServer := &http.Server{
		Handler:      handler,
		ReadTimeout:  rT,
		WriteTimeout: wT,
		Addr:         a,
	}

	s := &Server{
		server:          httpServer,
		notify:          make(chan error, 1),
		shutdownTimeout: _defaultShutdownTimeout,
	}

	s.start()

	return s
}

func (s *Server) start() {
	go func() {
		s.notify <- s.server.ListenAndServe()
		close(s.notify)
	}()
}

func (s *Server) Notify() <-chan error {
	return s.notify
}

func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()

	return s.server.Shutdown(ctx)
}
