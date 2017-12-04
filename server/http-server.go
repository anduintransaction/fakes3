package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/anduintransaction/fakes3/config"
)

// HTTPServer is a nice wrapper for net/http.Server that responses to SIGTERM and SIGINT
type HTTPServer struct {
	config     *config.HTTPConfig
	httpServer *http.Server
	signalChan chan os.Signal
	errorChan  chan error
}

// NewHTTPServer returns a new HTTPServer
func NewHTTPServer(config *config.HTTPConfig) *HTTPServer {
	return &HTTPServer{
		config:     config,
		signalChan: make(chan os.Signal, 1),
		errorChan:  make(chan error, 1),
	}
}

// Start starts a HTTP Server and register signal handlers to shutdown the server
func (s *HTTPServer) Start(handler http.Handler) {
	signal.Notify(s.signalChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		s.httpServer = &http.Server{
			Addr:    s.config.Addr,
			Handler: handler,
		}
		logrus.Infof("Starting HTTP server at %s", s.config.Addr)
		err := s.httpServer.ListenAndServe()
		s.errorChan <- err
	}()
}

// Wait waits for SIGTERM and SIGINT to shutdown the server
func (s *HTTPServer) Wait() error {
	select {
	case <-s.signalChan:
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		logrus.Infof("Shutting down HTTP Server at %s", s.httpServer.Addr)
		err := s.httpServer.Shutdown(ctx)
		if err != nil {
			logrus.Errorf("Error when shutting down server: %s", err)
		}
		return nil
	case err := <-s.errorChan:
		return err
	}
}
