package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

const (
	Red    = "\033[31m"
	Yellow = "\033[33m"
	Cyan   = "\033[36m"
	Reset  = "\033[0m"
)

type Logger struct {
	info  *log.Logger
	debug *log.Logger
	warn  *log.Logger
	err   *log.Logger
}

var (
	once sync.Once
	Log  *Logger
)

func init() {
	Log = &Logger{
		info:  log.New(os.Stdout, fmt.Sprintf("%s[INFO] %s", Cyan, Reset), log.LstdFlags),
		debug: log.New(os.Stdout, "[DEBUG] ", log.LstdFlags),
		warn:  log.New(os.Stdout, fmt.Sprintf("%s[WARN] %s", Yellow, Reset), log.LstdFlags),
		err:   log.New(os.Stderr, fmt.Sprintf("%s[ERROR] %s", Red, Reset), log.LstdFlags),
	}
}

func (l *Logger) callerInfo() string {
	// 2階層上の呼び出し元を取得（Info → Debug → 呼び出し元）
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		return ""
	}
	// ファイル名のみ抽出
	return fmt.Sprintf("%s:%d", filepath.Base(file), line)
}

func (l *Logger) Info(format string, args ...any) {
	prefix := fmt.Sprintf("[%s] ", l.callerInfo())
	l.info.Printf(prefix+format, args...)
}

func (l *Logger) Debug(format string, args ...any) {
	prefix := fmt.Sprintf("[%s] ", l.callerInfo())
	l.debug.Printf(prefix+format, args...)
}

func (l *Logger) Warn(format string, args ...any) {
	prefix := fmt.Sprintf("[%s] ", l.callerInfo())
	l.warn.Printf(prefix+format, args...)
}

func (l *Logger) Error(args ...any) {
	prefix := fmt.Sprintf("[%s] ", l.callerInfo())
	for i, arg := range args {
		if err, ok := arg.(error); ok {
			args[i] = err.Error()
		}
	}
	l.err.Println(append([]any{prefix}, args...)...)
}