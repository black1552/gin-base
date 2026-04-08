// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

// Package clickhouse implements database.Driver for ClickHouse database.
package clickhouse

import (
	"database/sql"
	"fmt"

	_ "github.com/ClickHouse/clickhouse-go/v2"

	"git.magicany.cc/black1552/gin-base/database"
)

// Driver is the driver for ClickHouse database.
type Driver struct {
	*database.Core
}

const (
	quoteChar = `"`
)

func init() {
	if err := database.Register("clickhouse", New()); err != nil {
		panic(err)
	}
}

// New creates and returns a driver that implements database.Driver for ClickHouse.
func New() database.Driver {
	return &Driver{}
}

// New creates and returns a database object for ClickHouse.
func (d *Driver) New(core *database.Core, node *database.ConfigNode) (database.DB, error) {
	return &Driver{
		Core: core,
	}, nil
}

// GetChars returns the security char for ClickHouse.
func (d *Driver) GetChars() (charLeft string, charRight string) {
	return quoteChar, quoteChar
}

// Open creates and returns an underlying sql.DB object for ClickHouse.
func (d *Driver) Open(config *database.ConfigNode) (*sql.DB, error) {
	var source string
	if config.Link != "" {
		source = config.Link
	} else {
		source = fmt.Sprintf("clickhouse://%s:%s@%s:%s/%s",
			config.User, config.Pass, config.Host, config.Port, config.Name)
	}
	return sql.Open("clickhouse", source)
}
