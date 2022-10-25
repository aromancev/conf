package main

import (
	"github.com/go-logr/logr"
	"github.com/rs/zerolog"
)

type Logger struct {
	log zerolog.Logger
}

func NewLogger(log zerolog.Logger) Logger {
	return Logger{
		log: log,
	}
}

// Init receives optional information about the logr library for LogSink
// implementations that need it.
func (l Logger) Init(info logr.RuntimeInfo) {

}

// Enabled tests whether this LogSink is enabled at the specified V-level.
// For example, commandline flags might be used to set the logging
// verbosity and disable some info logs.
func (l Logger) Enabled(level int) bool {
	return true
}

// Info logs a non-error message with the given key/value pairs as context.
// The level argument is provided for optional logging.  This method will
// only be called when Enabled(level) is true. See Logger.Info for more
// details.
func (l Logger) Info(level int, msg string, kvs ...interface{}) {
	var ev *zerolog.Event
	switch level {
	case 0:
		ev = l.log.Info()
	case 1:
		ev = l.log.Debug()
	default:
		ev = l.log.Trace()
	}

	for i := 0; i < len(kvs); i += 2 {
		ev.Interface(kvs[i].(string), kvs[i+1])
	}
	ev.Msg(msg)
}

// Error logs an error, with the given message and key/value pairs as
// context.  See Logger.Error for more details.
func (l Logger) Error(err error, msg string, kvs ...interface{}) {
	ev := l.log.Err(err)
	for i := 0; i < len(kvs); i++ {
		ev.Interface(kvs[i].(string), kvs[i+1])
	}
	ev.Msg(msg)
}

// WithValues returns a new LogSink with additional key/value pairs.  See
// Logger.WithValues for more details.
func (l Logger) WithValues(values ...interface{}) logr.LogSink {
	return l
}

// WithName returns a new LogSink with the specified name appended.  See
// Logger.WithName for more details.
func (l Logger) WithName(name string) logr.LogSink {
	return l
}
