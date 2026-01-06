package config

import (
	"fmt"
	"log"
	"os"
	"strings"
)

type Logger struct {
	logger *log.Logger
	name   string
	level  int
}

const (
	LevelDebug = iota
	LevelInfo
	LevelWarn
	LevelError
)

func NewLogger(name string) *Logger {
	l := &Logger{
		logger: log.New(os.Stdout, "", log.Ldate|log.Ltime),
		name:   name,
		level:  LevelInfo,
	}
	l.log("DEBUG", fmt.Sprintf("New Logger created with name: %s", name))
	return l
}

func (l *Logger) SetLevel(level string) {

	switch strings.ToLower(level) {
	case "debug":
		l.log("DEBUG", fmt.Sprintf("Logger set level %s for name%s", level, l.name))
		l.level = LevelDebug
	case "info":
		l.level = LevelInfo
	case "warn", "warning":
		l.level = LevelWarn
	case "error":
		l.level = LevelError
	}
}

func (l *Logger) Debug(msg string, args ...any) {
	if l.level <= LevelDebug {
		l.log("DEBUG", msg, args...)
	}
}
func (l *Logger) Info(msg string, args ...any) {
	if l.level <= LevelInfo {
		l.log("INFO", msg, args...)
	}
}
func (l *Logger) Warn(msg string, args ...any) {
	if l.level <= LevelWarn {
		l.log("WARN", msg, args...)
	}
}
func (l *Logger) Error(msg string, args ...any) {
	if l.level <= LevelError {
		l.log("ERROR", msg, args...)
	}
}
func (l *Logger) Fatal(msg string, args ...any) {
	l.log("FATAL", msg, args...)
	os.Exit(1)
}
func (l *Logger) log(level, msg string, args ...any) {
	var logArgs []any
	logArgs = append(logArgs, fmt.Sprintf("[%s] [%s]", l.name, level))
	logArgs = append(logArgs, msg)

	if len(args) > 0 {
		for i := 0; i < len(args); i += 2 {
			if i+1 < len(args) {
				logArgs = append(logArgs, fmt.Sprintf("%v=%v", args[i], args[i+1]))
			} else {
				logArgs = append(logArgs, args[i])
			}
		}
	}

	l.logger.Println(logArgs...)
}

func (l *Logger) Debugf(format string, args ...any) {
	if l.level <= LevelDebug {
		l.log("DEBUG", fmt.Sprintf(format, args...))
	}
}
func (l *Logger) Infof(format string, args ...any) {
	if l.level <= LevelInfo {
		l.log("INFO", fmt.Sprintf(format, args...))
	}
}
func (l *Logger) Warnf(format string, args ...any) {
	if l.level <= LevelWarn {
		l.log("WARN", fmt.Sprintf(format, args...))
	}
}
func (l *Logger) Errorf(format string, args ...any) {
	if l.level <= LevelError {
		l.log("ERROR", fmt.Sprintf(format, args...))
	}
}
func (l *Logger) Fatalf(format string, args ...any) {
	l.log("FATAL", fmt.Sprintf(format, args...))
	os.Exit(1)
}
