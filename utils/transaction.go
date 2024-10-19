package utils

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func mappingTxError(err, rbErr error) error {
	return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
}

func ExecTxPool(ctx context.Context, DB PGXPool, fn func(tx pgx.Tx) error, level ...pgx.TxIsoLevel) error {
	isolationLevel := pgx.ReadCommitted
	if len(level) > 0 {
		isolationLevel = level[0]
	}

	tx, err := DB.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: isolationLevel,
	})
	if err != nil {
		return err
	}

	err = fn(tx)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return mappingTxError(err, rbErr)
		}
		return err
	}

	return tx.Commit(ctx)
}
