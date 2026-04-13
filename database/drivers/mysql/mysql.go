// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

// Package mysql implements database.Driver, which supports operations for database MySQL.
package mysql

import (
	"git.magicany.cc/black1552/gin-base/database"
	"git.magicany.cc/black1552/gin-base/database/consts"
	_ "github.com/go-sql-driver/mysql"
)

// Driver is the driver for mysql database.
type Driver struct {
	*database.Core
}

const (
	quoteChar = "`"
)

func init() {
	var (
		err         error
		driverObj   = New()
		driverNames = consts.SliceStr{"mysql", "mariadb", "tidb"} // TODO remove mariadb and tidb in future versions.
	)
	for _, driverName := range driverNames {
		if err = database.Register(driverName, driverObj); err != nil {
			panic(err)
		}
	}
}

// New create and returns a driver that implements database.Driver, which supports operations for MySQL.
func New() database.Driver {
	return &Driver{}
}

// New creates and returns a database object for mysql.
// It implements the interface of database.Driver for extra database driver installation.
func (d *Driver) New(core *database.Core, node *database.ConfigNode) (database.DB, error) {
	return &Driver{
		Core: core,
	}, nil
}

// GetChars returns the security char for this type of database.
func (d *Driver) GetChars() (charLeft string, charRight string) {
	return quoteChar, quoteChar
}

// Migration returns a Migration instance for MySQL database operations.
func (d *Driver) Migration() *Migration {
	return NewMigration(d)
}

// GetMigration returns a Migration instance implementing the Migration interface.
func (d *Driver) GetMigration() database.Migration {
	return d.Migration()
}
