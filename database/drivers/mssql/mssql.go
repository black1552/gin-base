// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

// Package mssql implements database.Driver for Microsoft SQL Server.
package mssql

import (
	"database/sql"
	"fmt"

	_ "github.com/microsoft/go-mssqldb"

	"git.magicany.cc/black1552/gin-base/database"
)

// Driver is the driver for SQL Server database.
type Driver struct {
	*database.Core
}

const (
	quoteChar = `"`
)

func init() {
	if err := database.Register("mssql", New()); err != nil {
		panic(err)
	}
}

// New creates and returns a driver that implements database.Driver for SQL Server.
func New() database.Driver {
	return &Driver{}
}

// New creates and returns a database object for SQL Server.
func (d *Driver) New(core *database.Core, node *database.ConfigNode) (database.DB, error) {
	return &Driver{
		Core: core,
	}, nil
}

// GetChars returns the security char for SQL Server.
func (d *Driver) GetChars() (charLeft string, charRight string) {
	return quoteChar, quoteChar
}

// Open creates and returns an underlying sql.DB object for SQL Server.
func (d *Driver) Open(config *database.ConfigNode) (*sql.DB, error) {
	var source string
	if config.Link != "" {
		source = config.Link
	} else {
		source = fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s",
			config.User, config.Pass, config.Host, config.Port, config.Name)
	}
	return sql.Open("sqlserver", source)
}
