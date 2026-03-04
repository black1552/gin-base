package utils

import (
	"os"
	"path/filepath"

	"git.magicany.cc/black1552/gin-base/log"
	"github.com/gogf/gf/v2/container/garray"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/text/gstr"
)

func FileExists(path string) bool {
	return gfile.Exists(path)
}

func EmptyFile(path string) bool {
	return gfile.IsEmpty(path)
}

func GetResourceAllPath(path string) []string {
	if path == "" {
		path = "/"
	}
	basePath, _ := os.Getwd()
	filePath := filepath.Join(basePath, "resource")
	if path != "/" {
		filePath += filepath.Join(filePath, path)
	}
	f, err := os.Open(path)
	if err != nil {
		log.Error("open file error", err)
		panic(err)
	}
	paths, err := f.Readdirnames(-1)
	_ = f.Close()
	if err != nil {
		log.Error("read file error", err)
		panic(err)
	}
	pathArr := garray.NewStrArray()
	for _, v := range paths {
		if gstr.Contains(v, ".") {
			if path != "/" {
				pathArr.Append("/static/" + path + "/" + v)
			} else {
				pathArr.Append("/static/" + v)
			}
		} else {
			pathArr.Append(v)
		}
	}
	return pathArr.Slice()
}
