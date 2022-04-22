package http

import (
	"context"

	"github.com/yusufsyaifudin/ngendika/pkg/response"
	"github.com/yusufsyaifudin/ylog"
)

// Inject logger and response tracer at same time
func Inject(ctx context.Context, log *ylog.Tracer, resp response.Tracer) context.Context {
	return response.Inject(ylog.Inject(ctx, log), resp)
}
