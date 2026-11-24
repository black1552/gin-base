// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

// Package pgsql implements database.Driver, which supports operations for database PostgreSQL.
package pgsql

import (
	_ "github.com/lib/pq"

	"github.com/black1552/gin-base/database"
	"github.com/gogf/gf/v2/os/gctx"
)

// Driver is the driver for postgresql database.
type Driver struct {
	*database.Core
}

const (
	internalPrimaryKeyInCtx gctx.StrKey = "primary_key"
	defaultSchema           string      = "public"
	quoteChar               string      = `"`
)

func init() {
	if err := database.Register(`pgsql`, New()); err != nil {
		panic(err)
	}
}

// New create and returns a driver that implements database.Driver, which supports operations for PostgreSql.
func New() database.Driver {
	return &Driver{}
}

// New creates and returns a database object for postgresql.
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

// Migration returns a Migration instance for PostgreSQL database operations.
func (d *Driver) Migration() *Migration {
	return NewMigration(d)
}

// GetMigration returns a Migration instance implementing the Migration interface.
func (d *Driver) GetMigration() database.Migration {
	return d.Migration()
}
