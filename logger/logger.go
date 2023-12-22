package logger

import (
	"fmt"
	"log"
	"os"
)

var Log *Logger = NewLogger()

type LogLevel int

const (
	InfoLevel LogLevel = iota
	DebugLevel
	WarningLevel
	ErrorLevel
)

type Logger struct {
	info    *log.Logger
	debug   *log.Logger
	warning *log.Logger
	error   *log.Logger
}

func NewLogger() *Logger {
	return &Logger{
		info:    log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime),
		debug:   log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime),
		warning: log.New(os.Stdout, "WARNING: ", log.Ldate|log.Ltime),
		error:   log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime),
	}
}

func (l *Logger) Infof(format string, v ...interface{}) {
	l.Logf(InfoLevel, format, v...)
}

func (l *Logger) Info(message string) {
	l.Log(InfoLevel, message)
}

func (l *Logger) Debug(message string) {
	l.Log(DebugLevel, message)
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	l.Logf(DebugLevel, format, v...)
}

func (l *Logger) Warning(message string) {
	l.Log(WarningLevel, message)
}

func (l *Logger) Warningf(format string, v ...interface{}) {
	l.Logf(WarningLevel, format, v...)
}

func (l *Logger) Error(err error, message string) {
	l.Log(ErrorLevel, message+": "+err.Error())
}

func (l *Logger) Errorf(err error, format string, v ...interface{}) {
	l.Logf(ErrorLevel, format+": "+err.Error(), v...)
}

func (l *Logger) Log(level LogLevel, message string) {
	switch level {
	case InfoLevel:
		l.info.Println(message)
	case DebugLevel:
		l.debug.Println(message)
	case WarningLevel:
		l.warning.Println(message)
	case ErrorLevel:
		l.error.Println(message)
	default:
		l.info.Println(message)
	}
}

func (l *Logger) Logf(level LogLevel, format string, v ...interface{}) {
	formattedMessage := fmt.Sprintf(format, v...)
	l.Log(level, formattedMessage)
}
