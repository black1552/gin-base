// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

// Package sqlitecgo implements database.Driver, which supports operations for database SQLite.
//
// Note:
//  1. Using sqlitecgo is for building a 32-bit Windows operating system
//  2. You need to set the environment variable CGO_ENABLED=1 and make sure that GCC is installed
//     on your path. windows gcc: https://jmeubank.github.io/tdm-gcc/
package sqlitecgo

import (
	_ "github.com/mattn/go-sqlite3"

	"git.magicany.cc/black1552/gin-base/database"
)

// Driver is the driver for sqlite database.
type Driver struct {
	*database.Core
}

const (
	quoteChar = "`"
)

func init() {
	if err := database.Register(`sqlite`, New()); err != nil {
		panic(err)
	}
}

// New create and returns a driver that implements database.Driver, which supports operations for SQLite.
func New() database.Driver {
	return &Driver{}
}

// New creates and returns a database object for sqlite.
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
