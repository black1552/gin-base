// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

// Package drivers provides database drivers registration for our custom ORM.
// This package automatically registers all supported database drivers when imported.
// Supported databases: MySQL, MariaDB, PostgreSQL, SQL Server, Oracle, SQLite, ClickHouse
package drivers

import (
	// MySQL & MariaDB driver
	_ "git.magicany.cc/black1552/gin-base/database/drivers/mysql"

	// PostgreSQL driver
	_ "git.magicany.cc/black1552/gin-base/database/drivers/pgsql"

	// SQL Server driver
	_ "git.magicany.cc/black1552/gin-base/database/drivers/mssql"

	// Oracle driver
	_ "git.magicany.cc/black1552/gin-base/database/drivers/oracle"

	// SQLite driver (pure Go implementation, no CGO required)
	_ "git.magicany.cc/black1552/gin-base/database/drivers/sqlite"

	// ClickHouse driver
	_ "git.magicany.cc/black1552/gin-base/database/drivers/clickhouse"
)
