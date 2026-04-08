// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

// Package mysql implements database.Driver for MySQL database.
package mysql

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"

	"git.magicany.cc/black1552/gin-base/database"
)

// Driver is the driver for MySQL database.
type Driver struct {
	*database.Core
}

const (
	quoteChar = "`"
)

func init() {
	var (
		err       error
		driverObj = New()
	)
	// Register for both mysql and mariadb (compatible protocol)
	if err = database.Register("mysql", driverObj); err != nil {
		panic(err)
	}
	if err = database.Register("mariadb", driverObj); err != nil {
		panic(err)
	}
}

// New creates and returns a driver that implements database.Driver for MySQL.
func New() database.Driver {
	return &Driver{}
}

// New creates and returns a database object for MySQL.
// It implements the interface of database.Driver for extra database driver installation.
func (d *Driver) New(core *database.Core, node *database.ConfigNode) (database.DB, error) {
	return &Driver{
		Core: core,
	}, nil
}

// GetChars returns the security char for MySQL database.
func (d *Driver) GetChars() (charLeft string, charRight string) {
	return quoteChar, quoteChar
}

// Open creates and returns an underlying sql.DB object for MySQL.
func (d *Driver) Open(config *database.ConfigNode) (*sql.DB, error) {
	var (
		source   string
		username = config.User
		password = config.Pass
		protocol = "tcp"
		dbName   = config.Name
		params   = make([]string, 0)
	)

	if config.Extra != "" {
		params = append(params, config.Extra)
	}

	// Default params
	params = append(params, "loc=Local", "parseTime=true")

	source = fmt.Sprintf("%s:%s@%s(%s:%s)/%s?%s",
		username, password, protocol, config.Host, config.Port, dbName, strings.Join(params, "&"))

	return sql.Open("mysql", source)
}
