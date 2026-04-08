package database

import (
	"context"
)

type IDao interface {
	DB() DB
	TableName() string
	Columns() interface{}
	Group() string
	Ctx(ctx context.Context) *Model
	Transaction(ctx context.Context, f func(ctx context.Context, tx TX) error) (err error)
}
