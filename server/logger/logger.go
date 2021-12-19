package logger

import (
	"io"
	"log"
	"os"
)

// A logger contract describing how logging is supposed to behave in this app
type Logger interface {
	Err(v ...interface{})
	Errf(format string, v ...interface{})
	Warn(v ...interface{})
	Warnf(format string, v ...interface{})
	Info(v ...interface{})
	Infof(format string, v ...interface{})
}

type logger struct {
	err  *log.Logger
	warn *log.Logger
	info *log.Logger
}

func (l *logger) Err(v ...interface{}) {
	l.err.Println(v...)
}

func (l *logger) Errf(format string, v ...interface{}) {
	l.err.Printf(format, v...)
}

func (l *logger) Warn(v ...interface{}) {
	l.warn.Println(v...)
}

func (l *logger) Warnf(format string, v ...interface{}) {
	l.warn.Printf(format, v...)
}

func (l *logger) Info(v ...interface{}) {
	l.info.Println(v...)
}

func (l *logger) Infof(format string, v ...interface{}) {
	l.info.Printf(format, v...)
}

func New() *logger {
	return &logger{
		err:  newStdLogger(os.Stderr, "FAIL", false),
		warn: newStdLogger(os.Stdin, "WARN", false),
		info: newStdLogger(os.Stdin, "INFO", false),
	}
}

func newStdLogger(w io.Writer, prefix string, withFile bool) *log.Logger {
	flags := log.Ldate | log.Ltime
	if withFile {
		flags = log.Ldate | log.Ltime | log.Lshortfile
	}
	return log.New(w, prefix+" ", flags)
}
