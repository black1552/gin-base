// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

// Package oracle implements database.Driver for Oracle database.
package oracle

import (
	"database/sql"
	"fmt"

	_ "github.com/sijms/go-ora/v2"

	"git.magicany.cc/black1552/gin-base/database"
)

// Driver is the driver for Oracle database.
type Driver struct {
	*database.Core
}

const (
	quoteChar = `"`
)

func init() {
	if err := database.Register("oracle", New()); err != nil {
		panic(err)
	}
}

// New creates and returns a driver that implements database.Driver for Oracle.
func New() database.Driver {
	return &Driver{}
}

// New creates and returns a database object for Oracle.
func (d *Driver) New(core *database.Core, node *database.ConfigNode) (database.DB, error) {
	return &Driver{
		Core: core,
	}, nil
}

// GetChars returns the security char for Oracle.
func (d *Driver) GetChars() (charLeft string, charRight string) {
	return quoteChar, quoteChar
}

// Open creates and returns an underlying sql.DB object for Oracle.
func (d *Driver) Open(config *database.ConfigNode) (*sql.DB, error) {
	var source string
	if config.Link != "" {
		source = config.Link
	} else {
		source = fmt.Sprintf("oracle://%s:%s@%s:%s/%s",
			config.User, config.Pass, config.Host, config.Port, config.Name)
	}
	return sql.Open("oracle", source)
}
