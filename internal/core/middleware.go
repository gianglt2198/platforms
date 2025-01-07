package core

type Handler func(Request) (Response, error)
type Middleware func(Handler) Handler
