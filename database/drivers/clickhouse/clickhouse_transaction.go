// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

package clickhouse

import (
	"context"

	"git.magicany.cc/black1552/gin-base/database"
)

// Begin starts and returns the transaction object.
func (d *Driver) Begin(ctx context.Context) (tx database.TX, err error) {
	return nil, errUnsupportedBegin
}

// Transaction wraps the transaction logic using function `f`.
func (d *Driver) Transaction(ctx context.Context, f func(ctx context.Context, tx database.TX) error) error {
	return errUnsupportedTransaction
}
