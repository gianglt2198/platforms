package core

import (
	"context"
	oblogger "go-platform/internal/observability/logger"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type PlatformOption func(*Platform)
type Platform struct {
	// Name of the platform
	services    []Service
	middlewares []Middleware
	logger      *oblogger.Logger
	wg          sync.WaitGroup
}

func New(logger *oblogger.Logger, services ...Service) *Platform {
	return &Platform{
		services: services,
		logger:   logger,
	}
}

func (p *Platform) Use(middlewares ...Middleware) {
	p.middlewares = append(p.middlewares, middlewares...)
}

func (p *Platform) Run() error {
	// Initialize signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Initialize all services
	for _, svc := range p.services {
		if err := svc.Init(); err != nil {
			return err
		}
	}

	// Apply middlewares to all services
	for _, svc := range p.services {
		svc.Use(p.middlewares...)
	}

	// Start all services
	for _, svc := range p.services {
		p.wg.Add(1)
		go func(s Service) {
			defer p.wg.Done()
			if err := s.Start(); err != nil {
				p.logger.S.Errorf("Service %s failed: %v", s.Name(), err)
			}
		}(svc)
	}

	// Wait for shutdown signal
	<-sigChan
	return nil
}

func (p *Platform) Shutdown() error {
	p.logger.S.Info("Initiating graceful shutdown...")

	// Create a timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stop all services
	for _, svc := range p.services {
		if err := svc.Stop(ctx); err != nil {
			p.logger.S.Errorf("Error stopping service %s: %v", svc.Name(), err)
		}
	}

	// Wait for all goroutines to finish
	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		p.logger.S.Info("Graceful shutdown completed")
		return nil
	}
}
