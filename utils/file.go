package utils

import (
	"github.com/gogf/gf/v2/os/gfile"
)

func FileExists(path string) bool {
	return gfile.Exists(path)
}

func EmptyFile(path string) bool {
	return gfile.IsEmpty(path)
}
