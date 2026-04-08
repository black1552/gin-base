// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

// Package sqlite implements database.Driver for SQLite database.
package sqlite

import (
	"database/sql"

	_ "modernc.org/sqlite"

	"git.magicany.cc/black1552/gin-base/database"
)

// Driver is the driver for SQLite database.
type Driver struct {
	*database.Core
}

const (
	quoteChar = "`"
)

func init() {
	if err := database.Register("sqlite", New()); err != nil {
		panic(err)
	}
}

// New creates and returns a driver that implements database.Driver for SQLite.
func New() database.Driver {
	return &Driver{}
}

// New creates and returns a database object for SQLite.
func (d *Driver) New(core *database.Core, node *database.ConfigNode) (database.DB, error) {
	return &Driver{
		Core: core,
	}, nil
}

// GetChars returns the security char for SQLite.
func (d *Driver) GetChars() (charLeft string, charRight string) {
	return quoteChar, quoteChar
}

// Open creates and returns an underlying sql.DB object for SQLite.
func (d *Driver) Open(config *database.ConfigNode) (*sql.DB, error) {
	// For SQLite, use the Name field as the database file path
	dbName := config.Name
	if dbName == "" {
		dbName = ":memory:"
	}
	return sql.Open("sqlite", dbName)
}
