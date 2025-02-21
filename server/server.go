package mycore

import (
	"context"
	"time"
)

type ServerStatus string

const (
	ServerStatusStarting ServerStatus = "STARTING"
	ServerStatusRunning  ServerStatus = "RUNNING"
	ServerStatusStopping ServerStatus = "STOPPING"
	ServerStatusStopped  ServerStatus = "STOPPED"
	ServerStatusFailed   ServerStatus = "FAILED"
)

// HealthStatus represents the health check status
type HealthStatus struct {
	Status    ServerStatus   `json:"status"`
	Protocol  string         `json:"protocol"`
	Version   string         `json:"version"`
	Uptime    time.Duration  `json:"uptime"`
	StartTime time.Time      `json:"start_time"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// ErrorHandler handles server errors
type ErrorHandler func(err error) error

// Server represents the core server interface that all protocol implementations must satisfy
type Server interface {
	// Lifecycle method
	Start(context.Context) error
	Stop(context.Context) error
	Restart(context.Context) error

	// Service information
	Name() string
	Status() ServerStatus
	Protocol() string
	Version() string
	Health() (*HealthStatus, error)

	// Error hanlding
	OnError(ErrorHandler)
}

// StateChangeHandler handles server state changes
type StateChangeHandler func(serverName string, oldStatus, newStatus ServerStatus)

// ServerRegistry manages multiple protocol servers with dependency injection
type ServerRegistry interface {
	// Registration
	Register(server Server, deps ...any) error
	Unregister(name string) error

	// Information
	Get(name string) (Server, error)
	List() []Server
	Dependencies(name string) []any

	// Events
	OnServerStateChange(handler StateChangeHandler)

	// Lifecycle
	StartAll(ctx context.Context) error
	StopAll(ctx context.Context) error
}
