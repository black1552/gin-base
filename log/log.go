package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime/debug"
	"strings"

	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gtime"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	logPath string
	sysLog  *log.Logger
)

func init() {
	logPath = gfile.Join(gfile.Pwd(), "logs")
	filePath := gfile.Join(logPath, fmt.Sprintf("log-%s.log", gtime.Date()))
	fileLogger := &lumberjack.Logger{
		Filename:   filePath,
		MaxSize:    2,     // 单个文件最大 10MB
		MaxBackups: 5,     // 最多保留 5 个备份
		MaxAge:     30,    // 保留 30 天
		Compress:   false, // 启用压缩
	}
	// 创建 MultiWriter 实现同时输出到文件和终端
	multiWriter := io.MultiWriter(fileLogger, os.Stdout)
	sysLog = log.New(multiWriter, "", log.LstdFlags)
}

func Info(v ...any) {
	sysLog.SetPrefix("[INFO] ")
	sysLog.Println(fmt.Sprint(v...))
}
func Error(v ...any) {
	sysLog.SetPrefix("[ERROR] ")
	msg := fmt.Sprint(v...)
	sysLog.Println(msg, strings.TrimSpace(string(debug.Stack())))
}
func Warn(v ...any) {
	sysLog.SetPrefix("[WARN] ")
	sysLog.Println(fmt.Sprint(v...))
}
func Debug(v ...any) {
	sysLog.SetPrefix("[DEBUG] ")
	sysLog.Println(fmt.Sprint(v...))
}
