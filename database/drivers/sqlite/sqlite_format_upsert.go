// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

package sqlite

import (
	"fmt"

	"github.com/black1552/gin-base/database"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
)

// FormatUpsert returns SQL clause of type upsert for SQLite.
// For example: ON CONFLICT (id) DO UPDATE SET ...
func (d *Driver) FormatUpsert(columns []string, list database.List, option database.DoInsertOption) (string, error) {
	if len(option.OnConflict) == 0 {
		return "", gerror.NewCode(
			gcode.CodeMissingParameter, `Please specify conflict columns`,
		)
	}

	var onDuplicateStr string
	if option.OnDuplicateStr != "" {
		onDuplicateStr = option.OnDuplicateStr
	} else if len(option.OnDuplicateMap) > 0 {
		for k, v := range option.OnDuplicateMap {
			if len(onDuplicateStr) > 0 {
				onDuplicateStr += ","
			}
			switch v.(type) {
			case database.Raw, *database.Raw:
				onDuplicateStr += fmt.Sprintf(
					"%s=%s",
					d.Core.QuoteWord(k),
					v,
				)
			case database.Counter, *database.Counter:
				var counter database.Counter
				switch value := v.(type) {
				case database.Counter:
					counter = value
				case *database.Counter:
					counter = *value
				}
				operator, columnVal := "+", counter.Value
				if columnVal < 0 {
					operator, columnVal = "-", -columnVal
				}
				onDuplicateStr += fmt.Sprintf(
					"%s=EXCLUDED.%s%s%s",
					d.QuoteWord(k),
					d.QuoteWord(counter.Field),
					operator,
					gconv.String(columnVal),
				)
			default:
				onDuplicateStr += fmt.Sprintf(
					"%s=EXCLUDED.%s",
					d.Core.QuoteWord(k),
					d.Core.QuoteWord(gconv.String(v)),
				)
			}
		}
	} else {
		for _, column := range columns {
			// If it's SAVE operation, do not automatically update the creating time.
			if d.Core.IsSoftCreatedFieldName(column) {
				continue
			}
			if len(onDuplicateStr) > 0 {
				onDuplicateStr += ","
			}
			onDuplicateStr += fmt.Sprintf(
				"%s=EXCLUDED.%s",
				d.Core.QuoteWord(column),
				d.Core.QuoteWord(column),
			)
		}
	}

	conflictKeys := gstr.Join(option.OnConflict, ",")

	return fmt.Sprintf("ON CONFLICT (%s) DO UPDATE SET ", conflictKeys) + onDuplicateStr, nil
}
