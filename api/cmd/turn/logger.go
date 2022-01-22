package main

import (
	"github.com/pion/logging"
	"github.com/rs/zerolog"
)

type LoggerFactory func(scope string) logging.LeveledLogger

func (f LoggerFactory) NewLogger(scope string) logging.LeveledLogger {
	return f(scope)
}

type Logger struct {
	log zerolog.Logger
}

func NewLogger(log zerolog.Logger) *Logger {
	return &Logger{
		log: log,
	}
}

func (l Logger) Trace(msg string) {
	l.log.Trace().Msg(msg)
}

func (l Logger) Tracef(format string, args ...interface{}) {
	l.log.Trace().Msgf(format, args...)
}

func (l Logger) Debug(msg string) {
	l.log.Debug().Msg(msg)
}

func (l Logger) Debugf(format string, args ...interface{}) {
	l.log.Debug().Msgf(format, args...)
}

func (l Logger) Info(msg string) {
	l.log.Info().Msg(msg)
}

func (l Logger) Infof(format string, args ...interface{}) {
	l.log.Info().Msgf(format, args...)
}

func (l Logger) Warn(msg string) {
	l.log.Warn().Msg(msg)
}

func (l Logger) Warnf(format string, args ...interface{}) {
	l.log.Warn().Msgf(format, args...)
}

func (l Logger) Error(msg string) {
	l.log.Error().Msg(msg)
}

func (l Logger) Errorf(format string, args ...interface{}) {
	l.log.Error().Msgf(format, args...)
}
