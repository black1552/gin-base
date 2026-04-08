// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

// Package oracle implements database.Driver, which supports operations for database Oracle.
package oracle

import (
	"git.magicany.cc/black1552/gin-base/database"
)

// Driver is the driver for oracle database.
type Driver struct {
	*database.Core
}

const (
	rowNumberAliasForSelect = `ROW_NUMBER__`
	quoteChar               = `"`
)

func init() {
	if err := database.Register(`oracle`, New()); err != nil {
		panic(err)
	}
}

// New create and returns a driver that implements database.Driver, which supports operations for Oracle.
func New() database.Driver {
	return &Driver{}
}

// New creates and returns a database object for oracle.
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
