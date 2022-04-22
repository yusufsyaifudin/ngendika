package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"github.com/yusufsyaifudin/ylog"
)

type Query struct {
	sqlConn sqlx.QueryerContext `validate:"required"`
}

func (p *Query) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	ylog.Info(ctx, "QueryContext", ylog.KV("query", formatQuery(query, args...)))
	return p.sqlConn.QueryContext(ctx, query, args...)
}

func (p *Query) QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error) {
	ylog.Info(ctx, "QueryxContext", ylog.KV("query", formatQuery(query, args...)))
	return p.sqlConn.QueryxContext(ctx, query, args...)
}

func (p *Query) QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row {
	ylog.Info(ctx, "QueryRowxContext", ylog.KV("query", formatQuery(query, args...)))
	return p.sqlConn.QueryRowxContext(ctx, query, args...)
}

// formatQuery formatting query
func formatQuery(q string, args ...interface{}) string {
	for i := 0; i < len(args); i++ {
		positionalArg := fmt.Sprint("$", i+1) // start from $1

		arg := fmt.Sprintf("'%v'", args[i])
		if v, ok := args[i].(time.Time); ok {
			arg = fmt.Sprintf("'%s'", v.Format(time.RFC3339Nano))
		}

		if v, ok := args[i].(bool); ok {
			arg = fmt.Sprintf("%t", v)
		}

		q = strings.ReplaceAll(q, positionalArg, arg)
	}

	return q
}

var _ sqlx.QueryerContext = (*Query)(nil)

func WrapWithLog(conn *sqlx.DB) (*Query, error) {
	instance := &Query{
		sqlConn: conn,
	}

	err := validator.New().Struct(instance)
	return instance, err
}
