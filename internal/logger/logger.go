package logger

import (
	"log"
	"os"
)

type Logger struct {
	l *log.Logger
}

func New() *Logger {
	return &Logger {
		l: log.New(os.Stderr, "[argusenum] ", log.LstdFlags),
	}
}

func (lg *Logger) Infof(format string, v ...interface{}) {
	lg.l.Printf("INFO: "+format, v...)
}

func (lg *Logger) Warnf(format string, v ...interface{}) {
	lg.l.Printf("WARN: "+format, v...)
}

func (lg *Logger) Errorf(format string, v ...interface{}) {
	lg.l.Printf("ERROR: "+format, v...)
}

func (lg *Logger) Debugf(format string, v ...interface{}) {
	lg.l.Printf("DEBUG: "+format, v...)
}