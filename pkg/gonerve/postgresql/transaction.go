package postgresql

import (
	"context"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"

	"stockpilot/pkg/gonerve/errors"
	"stockpilot/pkg/gonerve/logging"
)

func (r *Repository) WithTx(ctx context.Context, f func(ctx context.Context, tx pgx.Tx) error) (err error) {
	if err = r.Locked(); err != nil {
		return err
	}

	var tx pgx.Tx
	tx, err = r.Conn.Begin(ctx)
	if err != nil {
		return errors.Wrap(err, "begin tx")
	}

	defer func() {
		if err == nil {
			err = r.Locked()
		}

		if err != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				logging.Error(ctx, "rollback tx", zap.Error(rbErr))
			}
			return
		}

		if cmErr := tx.Commit(ctx); cmErr != nil {
			err = errors.Wrap(cmErr, "commit tx")
		}
	}()

	return f(ctx, tx)
}
