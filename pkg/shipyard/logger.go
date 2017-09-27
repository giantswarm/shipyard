package shipyard

import (
	"log"
	"strings"
)

var Log Logger = &StandardLogger{}

// Logger wraps common log and Gingko logging.
type Logger interface {
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Log(args ...interface{})
	Logf(format string, args ...interface{})
	LogWithPrefix(prefix string, str string)
}

type StandardLogger struct{}

func (*StandardLogger) Fatal(args ...interface{}) {
	log.Fatal(args...)
}

func (*StandardLogger) Fatalf(format string, args ...interface{}) {
	log.Fatalf(format, args...)
}

func (*StandardLogger) Log(args ...interface{}) {
	log.Print(args...)
}

func (*StandardLogger) Logf(format string, args ...interface{}) {
	log.Printf(format, args...)
}

func (l *StandardLogger) LogWithPrefix(prefix string, str string) {
	LogWithPrefix(log.Printf, prefix, str)
}

func LogWithPrefix(lf func(format string, args ...interface{}), prefix string, str string) {
	lines := strings.Split(str, "\n")
	for _, line := range lines {
		lf("%v | %v", prefix, line)
	}
}
