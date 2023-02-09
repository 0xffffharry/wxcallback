package core

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"wxcallback/lib/log"
)

func (c *Config) NewServer(option *ServerOption) *Server {
	s := &Server{
		config: *c,
	}
	if option.Logger != nil {
		s.logger = option.Logger
	} else {
		s.logger = log.NewLogger(nil, nil)
	}
	if option.Context != nil {
		s.context = option.Context
	} else {
		s.context = context.Background()
	}
	return s
}

func (s *Server) RunWithContext(ctx context.Context) {
	if ctx != nil {
		s.context = ctx
	}
	s.logger.Info("Server", "Starting...")
	defer s.logger.Info("Server", "Closed")
	switch s.config.Mode {
	case "port":
		s.logger.Info("Server", fmt.Sprintf("Mode: %s", s.config.Mode))
		s.runForPort()
	case "path":
		s.logger.Info("Server", fmt.Sprintf("Mode: %s", s.config.Mode))
		s.runForPath()
	}
}

func (s *Server) Run() {
	s.RunWithContext(context.Background())
}

func (s *Server) runForPort() {
	wg := sync.WaitGroup{}
	for _, service := range s.config.Service {
		wg.Add(1)
		go func(service Service) {
			defer wg.Done()
			httpServer := &http.Server{}
			httpServer.Addr = s.config.Listen
			httpServerMux := &http.ServeMux{}
			s.logger.Info("Server", fmt.Sprintf("Add Service, Listen: %s", service.Listen))
			flag := fmt.Sprintf("Server - %s: %s", "Listen", service.Listen)
			httpServerMux.HandleFunc(service.Path, func(w http.ResponseWriter, r *http.Request) {
				s.handler(w, r, flag, service)
			})
			httpServer.Handler = httpServerMux
			go func() {
				<-s.context.Done()
				httpServer.Close()
			}()
			if err := httpServer.ListenAndServe(); err != nil {
				if err != http.ErrServerClosed {
					s.logger.Error("Server", fmt.Sprintf("Service Listen: %s, Close Fail: %s", service.Listen, err))
				}
			}
		}(service)
	}
	wg.Wait()
}

func (s *Server) runForPath() {
	httpServer := &http.Server{}
	httpServer.Addr = s.config.Listen
	httpServerMux := &http.ServeMux{}
	rootPathTag := false
	for _, service := range s.config.Service {
		if service.Path == "/" {
			rootPathTag = true
		}
		s.logger.Info("Server", fmt.Sprintf("Add Service, Path: %s", service.Path))
		flag := fmt.Sprintf("Server - %s: %s", "Path", service.Path)
		httpServerMux.HandleFunc(service.Path, func(w http.ResponseWriter, r *http.Request) {
			s.handler(w, r, flag, service)
		})
	}
	if !rootPathTag {
		httpServerMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})
	}
	httpServer.Handler = httpServerMux
	go func() {
		<-s.context.Done()
		httpServer.Close()
	}()
	if err := httpServer.ListenAndServe(); err != nil {
		if err != http.ErrServerClosed {
			s.logger.Error("Server", fmt.Sprintf("Close Fail: %s", err))
		}
	}
}
