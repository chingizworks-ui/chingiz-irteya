package query

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"

	"stockpilot/pkg/gonerve/errors"
	"stockpilot/pkg/gonerve/logging"
	"stockpilot/pkg/gonerve/postgresql"
)

type dbConn interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

var replacer = strings.NewReplacer("\n", " ", "\t", " ")

func GetAll[T any](ctx context.Context, c dbConn, q string, args ...any) ([]T, error) {
	logging.Debug(ctx, "executing: ", zap.String("query", replacer.Replace(q)), zap.String("args", fmt.Sprintf("%v", args)))
	return GetAllNoLog[T](ctx, c, q, args...)
}

func GetAllNoLog[T any](ctx context.Context, c dbConn, q string, args ...any) ([]T, error) {
	rows, err := c.Query(ctx, q, args...)
	if err != nil {
		return nil, errors.Wrap(err, "database query failed")
	}

	return pgx.CollectRows(rows, pgx.RowToStructByNameLax[T])
}

func GetOne[T any](ctx context.Context, c dbConn, q string, args ...any) (*T, error) {
	res, err := GetAll[T](ctx, c, q, args...)
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, errors.ErrNotFound
	}
	return &res[0], nil
}

func Exec(ctx context.Context, c dbConn, q string, args ...any) error {
	logging.Debug(ctx, "executing: ", zap.String("query", q), zap.String("args", fmt.Sprintf("%v", args)))
	r, err := c.Exec(ctx, q, args...)
	if err != nil {
		return errors.Wrap(err, "database query failed")
	}
	if r.RowsAffected() == 0 {
		return errors.ErrNotFound
	}
	return nil
}

type ConverterWithError[T, R any] func(t T) (R, error)

func SelectOneWithConverterError[T, R any](ctx context.Context, conn postgresql.Connection, q string, converter ConverterWithError[T, R], args ...any) (R, error) {
	v, err := GetOne[T](ctx, conn, q, args...)
	if err != nil {
		return *new(R), err
	}
	res, err := converter(*v)
	if err != nil {
		return *new(R), errors.Wrap(err, "convert")
	}
	return res, nil
}
