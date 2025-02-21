package adapters

import (
	"context"

	"github.com/gianglt2198/platforms/internal/core"

	"google.golang.org/grpc"
)

type GRPCRequest struct {
	ctx     context.Context
	method  string
	payload interface{}
	meta    map[string]interface{}
}

func (r *GRPCRequest) Context() context.Context         { return r.ctx }
func (r *GRPCRequest) Protocol() string                 { return "grpc" }
func (r *GRPCRequest) Method() string                   { return r.method }
func (r *GRPCRequest) Path() string                     { return r.method }
func (r *GRPCRequest) Payload() interface{}             { return r.payload }
func (r *GRPCRequest) Metadata() map[string]interface{} { return r.meta }

type GRPCResponse struct {
	payload interface{}
	err     error
	meta    map[string]interface{}
}

func (r *GRPCResponse) Payload() interface{}             { return r.payload }
func (r *GRPCResponse) Error() error                     { return r.err }
func (r *GRPCResponse) Metadata() map[string]interface{} { return r.meta }

func UnaryServerInterceptor(middlewares ...core.Middleware) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Convert to core.Request
		coreReq := &GRPCRequest{
			ctx:     ctx,
			method:  info.FullMethod,
			payload: req,
		}

		// Chain middlewares
		var h core.Handler = func(r core.Request) (core.Response, error) {
			resp, err := handler(r.Context(), r.Payload())
			return &GRPCResponse{payload: resp}, err
		}

		for i := len(middlewares) - 1; i >= 0; i-- {
			h = middlewares[i](h)
		}

		resp, err := h(coreReq)
		if err != nil {
			return nil, err
		}

		return resp.Payload(), nil
	}
}
