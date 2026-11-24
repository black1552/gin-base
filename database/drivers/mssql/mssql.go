// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

// Package mssql implements database.Driver, which supports operations for MSSQL.
package mssql

import (
	_ "github.com/microsoft/go-mssqldb"

	"github.com/black1552/gin-base/database"
)

// Driver is the driver for SQL server database.
type Driver struct {
	*database.Core
}

const (
	rowNumberAliasForSelect = `ROW_NUMBER__`
	quoteChar               = `"`
)

func init() {
	if err := database.Register(`mssql`, New()); err != nil {
		panic(err)
	}
}

// New create and returns a driver that implements database.Driver, which supports operations for Mssql.
func New() database.Driver {
	return &Driver{}
}

// New creates and returns a database object for SQL server.
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

// Migration returns a Migration instance for MSSQL database operations.
func (d *Driver) Migration() *Migration {
	return NewMigration(d)
}

// GetMigration returns a Migration instance implementing the Migration interface.
func (d *Driver) GetMigration() database.Migration {
	return d.Migration()
}
