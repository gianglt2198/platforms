package mydatabase

import (
	"context"
	"log/slog"

	myerrors "github.com/gianglt2198/platforms/internal/errors"
	"gorm.io/gorm"
)

type (
	Transaction[T any] struct {
		db *gorm.DB
	}
	TransactionIf[T any] interface {
		Execute(context.Context, func() (*T, *myerrors.AppError)) (*T, *myerrors.AppError)
	}
)

func NewTransaction[T any](db *gorm.DB) *Transaction[T] {
	return &Transaction[T]{db}
}

func (t *Transaction[T]) Execute(ctx context.Context, f func(context.Context) (*T, *myerrors.AppError)) (*T, *myerrors.AppError) {

	slog.Info("[Transaction]*****Begin*****")
	tx := t.db.Begin()

	defer func() {
		if r := recover(); r != nil {
			slog.Info("[Transaction]--rollback due to runtime panic")
			tx.Rollback()
		}
		slog.Info("[Transaction]*****End*****")
	}()

	if err := tx.Error; err != nil {
		slog.Error("[Transaction]fail to begin tran", slog.Any("err", err))
		return nil, myerrors.QueryInvalid(err.Error())
	}

	slog.Info("[Transaction]--executing...")
	ctx = context.WithValue(ctx, KEY_CURRENT_TRAN, tx)

	result, aerr := f(ctx)

	if aerr != nil {
		slog.Info("[Transaction]--rollback to release locked by tran", slog.Any("aerr", aerr))
		tx.Rollback()
		return nil, aerr
	}

	slog.Info("[Transaction]--commit tran")
	err := tx.Commit().Error
	if err != nil {
		return nil, myerrors.QueryInvalid(err.Error())
	}

	return result, nil
}
