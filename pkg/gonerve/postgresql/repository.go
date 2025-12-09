package postgresql

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"stockpilot/pkg/gonerve/errors"
)

type Option func(r *Repository)

func WithListenNotifications(v bool) Option {
	return func(r *Repository) {
		r.listen = v
	}
}

type Repository struct {
	Conn     Connection
	isLocked atomic.Bool
	TxConn   bool
	listen   bool
}

func NewRepository(ctx context.Context, connString string, opts ...Option) (*Repository, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, errors.Wrap(err, "parse config")
	}

	connectCtx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	conn, err := pgxpool.NewWithConfig(connectCtx, config)
	if err != nil {
		return nil, err
	}
	c, err := conn.Acquire(connectCtx)
	if err != nil {
		return nil, errors.Wrap(err, "acquire connection")
	}
	c.Release()

	r := &Repository{
		Conn: conn,
	}

	for _, opt := range opts {
		opt(r)
	}

	return r, nil
}

type Connection interface {
	Ping(ctx context.Context) error
	Exec(ctx context.Context, sql string, arguments ...any) (commandTag pgconn.CommandTag, err error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)

	Begin(ctx context.Context) (pgx.Tx, error)
	Close()
}

func (r *Repository) Close() {
	r.Conn.Close()
}

func (r *Repository) Locked() error {
	if !r.isLocked.CompareAndSwap(false, false) {
		return errors.New("db is locked")
	}
	return nil
}
