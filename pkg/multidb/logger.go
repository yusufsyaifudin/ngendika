package multidb

import (
	"context"
	sqldblogger "github.com/simukti/sqldb-logger"
	"github.com/yusufsyaifudin/ylog"
)

type QueryLogger struct{}

func (q *QueryLogger) Log(ctx context.Context, level sqldblogger.Level, msg string, data map[string]interface{}) {
	ylog.Debug(ctx, msg, ylog.KV("level", level), ylog.KV("sql", data))
}

var _ sqldblogger.Logger = (*QueryLogger)(nil)
