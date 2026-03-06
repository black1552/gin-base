package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gtime"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	logPath string
	sysLog  *log.Logger
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
)

// 正则表达式匹配 ANSI 颜色码
var ansiColorRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// stripAnsiColors 去除字符串中的 ANSI 颜色码
func stripAnsiColors(s string) string {
	return ansiColorRegex.ReplaceAllString(s, "")
}

// logWriter 自定义 writer，用于分别处理控制台和文件输出
type logWriter struct {
	console io.Writer
	file    io.Writer
}

func (w *logWriter) Write(p []byte) (n int, err error) {
	// 控制台输出保留颜色
	_, err = w.console.Write(p)
	if err != nil {
		return 0, err
	}
	// 文件输出去除颜色码
	colorless := stripAnsiColors(string(p))
	_, err = w.file.Write([]byte(colorless))
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

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
	// 使用自定义 writer 实现控制台带颜色、文件无颜色的输出
	multiWriter := &logWriter{
		console: os.Stdout,
		file:    fileLogger,
	}
	sysLog = log.New(multiWriter, "", 0)
}

func Info(v ...any) {
	sysLog.SetPrefix(fmt.Sprintf("[%s] %s[INFO]%s ", time.Now().Format("2006-01-02 15:04:05"), Green, Reset))
	sysLog.Println(fmt.Sprint(v...))
}
func Error(v ...any) {
	sysLog.SetPrefix(fmt.Sprintf("[%s] %s[ERROR]%s ", time.Now().Format("2006-01-02 15:04:05"), Red, Reset))
	msg := fmt.Sprint(v...)
	sysLog.Println(msg, strings.TrimSpace(string(debug.Stack())))
}
func Warn(v ...any) {
	sysLog.SetPrefix(fmt.Sprintf("[%s] %s[WARN]%s ", time.Now().Format("2006-01-02 15:04:05"), Yellow, Reset))
	sysLog.Println(fmt.Sprint(v...))
}
func Debug(v ...any) {
	sysLog.SetPrefix(fmt.Sprintf("[%s] %s[DEBUG]%s ", time.Now().Format("2006-01-02 15:04:05"), Blue, Reset))
	sysLog.Println(fmt.Sprint(v...))
}
