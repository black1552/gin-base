// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

package sqlite

import (
	"context"

	"git.magicany.cc/black1552/gin-base/database"
	"github.com/gogf/gf/v2/text/gstr"
)

// DoFilter deals with the sql string before commits it to underlying sql driver.
func (d *Driver) DoFilter(
	ctx context.Context, link database.Link, sql string, args []any,
) (newSql string, newArgs []any, err error) {
	// Special insert/ignore operation for sqlite.
	switch {
	case gstr.HasPrefix(sql, database.InsertOperationIgnore):
		sql = "INSERT OR IGNORE" + sql[len(database.InsertOperationIgnore):]

	case gstr.HasPrefix(sql, database.InsertOperationReplace):
		sql = "INSERT OR REPLACE" + sql[len(database.InsertOperationReplace):]
	}
	return d.Core.DoFilter(ctx, link, sql, args)
}
