package logger

import (
    "log"
)

type Logger struct {
    prefix string
}

func New() *Logger {
    return &Logger{prefix: "[ArgusEnum] "}
}

func (l *Logger) Info(msg string) {
    log.Println(l.prefix + msg)
}

func (l *Logger) Infof(format string, a ...interface{}) {
    log.Printf(l.prefix+format, a...)
}

func (l *Logger) Error(msg string) {
    log.Println(l.prefix + "[ERROR] " + msg)
}

func (l *Logger) Errorf(format string, a ...interface{}) {
    log.Printf(l.prefix+"[ERROR] "+format, a...)
}

func (l *Logger) Debug(msg string) {
    log.Println(l.prefix + "[DEBUG] " + msg)
}

func (l *Logger) Debugf(format string, a ...interface{}) {
    log.Printf(l.prefix+"[DEBUG] "+format, a...)
}
