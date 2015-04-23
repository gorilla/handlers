package log

import (
	"fmt"
	"io"
	"os"
	"sync"
)

type Level int8

type Log struct {
	Formatter
	mu        sync.Mutex
	threshold Level
	writer    io.Writer
}

const (
	None Level = iota - 1
	Fatal
	Error
	Warning
	Notice
	Info
	Debug
)

var (
	levelStrings = []string{
		"FATAL",
		"ERROR",
		"WARNING",
		"NOTICE",
		"INFO",
		"DEBUG",
	}
)

func New(writer io.Writer, threshold Level) *Log {
	hostname, _ := os.Hostname()
	return &Log{
		Formatter: &DefaultFormat{
			hostname: hostname,
			pid:      os.Getpid(),
			tag:      os.Args[0],
		},
		threshold: threshold,
		writer:    writer,
	}
}

func (log *Log) Debug(args ...interface{}) {
	log.write(Debug, args)
}

func (log *Log) Debugf(format string, args ...interface{}) {
	log.write(Debug, fmt.Sprintf(format, args))
}

func (log *Log) Info(args ...interface{}) {
	log.write(Info, args)
}

func (log *Log) Infof(format string, args ...interface{}) {
	log.write(Info, fmt.Sprintf(format, args))
}

func (log *Log) Notice(args ...interface{}) {
	log.write(Notice, args)
}

func (log *Log) Noticef(format string, args ...interface{}) {
	log.write(Notice, fmt.Sprintf(format, args))
}

func (log *Log) Warning(args ...interface{}) {
	log.write(Warning, args)
}

func (log *Log) Warningf(format string, args ...interface{}) {
	log.write(Warning, fmt.Sprintf(format, args))
}

func (log *Log) Error(args ...interface{}) {
	log.write(Error, args)
}

func (log *Log) Errorf(format string, args ...interface{}) {
	log.write(Error, fmt.Sprintf(format, args))
}

func (log *Log) Fatal(args ...interface{}) {
	log.write(Fatal, args)
	os.Exit(1)
}

func (log *Log) Fatalf(format string, args ...interface{}) {
	log.write(Fatal, fmt.Sprintf(format, args))
	os.Exit(1)
}

func (log *Log) SetFormatter(f Formatter) {
	log.Formatter = f
}

func (log *Log) write(level Level, args ...interface{}) {
	if level > log.threshold {
		return
	}

	log.mu.Lock()
	defer log.mu.Unlock()

	fmt.Fprint(log.writer, log.Format(level, args))
}

func (level Level) String() string {
	return levelStrings[level]
}
