package log

import (
	"fmt"
	"time"
)

type Formatter interface {
	Format(Level, ...interface{}) string
}

type DefaultFormat struct {
	hostname string
	pid      int
	tag      string
}

func (f *DefaultFormat) Format(level Level, args ...interface{}) string {
	timestamp := time.Now().Format(time.RFC3339)

	return fmt.Sprintf("%s %s %s[%d]: %s %v\n",
		timestamp, f.hostname, f.tag, f.pid, level, args)
}

func (f *DefaultFormat) SetTag(t string) {
	f.tag = t
}
