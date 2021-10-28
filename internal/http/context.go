package http

import (
	"context"

	"github.com/yusufsyaifudin/ngendika/pkg/logger"
	"github.com/yusufsyaifudin/ngendika/pkg/response"
)

// Inject logger and response tracer at same time
func Inject(ctx context.Context, log logger.Tracer, resp response.Tracer) context.Context {
	return response.Inject(logger.Inject(ctx, log), resp)
}
