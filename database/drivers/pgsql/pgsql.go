// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

// Package pgsql implements database.Driver for PostgreSQL database.
package pgsql

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"

	"git.magicany.cc/black1552/gin-base/database"
)

// Driver is the driver for PostgreSQL database.
type Driver struct {
	*database.Core
}

const (
	quoteChar = `"`
)

func init() {
	if err := database.Register("pgsql", New()); err != nil {
		panic(err)
	}
}

// New creates and returns a driver that implements database.Driver for PostgreSQL.
func New() database.Driver {
	return &Driver{}
}

// New creates and returns a database object for PostgreSQL.
// It implements the interface of database.Driver for extra database driver installation.
func (d *Driver) New(core *database.Core, node *database.ConfigNode) (database.DB, error) {
	return &Driver{
		Core: core,
	}, nil
}

// GetChars returns the security char for PostgreSQL database.
func (d *Driver) GetChars() (charLeft string, charRight string) {
	return quoteChar, quoteChar
}

// Open creates and returns an underlying sql.DB object for PostgreSQL.
func (d *Driver) Open(config *database.ConfigNode) (*sql.DB, error) {
	var (
		source   string
		username = config.User
		password = config.Pass
		host     = config.Host
		port     = config.Port
		dbName   = config.Name
	)

	source = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, username, password, dbName)

	if config.Extra != "" {
		source += " " + config.Extra
	}

	return sql.Open("postgres", source)
}
