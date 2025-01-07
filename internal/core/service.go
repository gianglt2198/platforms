package core

import "context"

type Request interface {
	Context() context.Context
	Protocol() string
	Method() string
	Path() string
	Payload() interface{}
	Metadata() map[string]interface{}
}

type Response interface {
	Payload() interface{}
	Error() error
	Metadata() map[string]interface{}
}

type Service interface {
	Init() error
	Start() error
	Stop(context.Context) error
	Name() string
	Use(...Middleware)
}
