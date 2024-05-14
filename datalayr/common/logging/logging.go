package logging

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/Layr-Labs/datalayr/common/middleware/correlation"
	glogging "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/rs/zerolog"
)

type Logger struct {
	*zerolog.Logger
}

// For unit tests
func GetNoopLogger() *Logger {
	logger := zerolog.Nop()
	return &Logger{&logger}
}

// Sublogger creates a child logger with the "loc" field added.
// Primarily used as a way to distinguish between different components of
// our system.
func (l *Logger) Sublogger(loc string) *Logger {
	sublogger := l.With().Str("loc", loc).Logger()
	return &Logger{Logger: &sublogger}
}

// SubloggerCtx adds a correlation ID as a field to the logger.
func (l *Logger) SubloggerId(ctx context.Context) *Logger {
	fields := glogging.ExtractFields(ctx)
	vals := make(map[string]interface{}, len(fields)/2)
	for i := 0; i < len(fields); i += 2 {
		vals[fields[i]] = fields[i+1]
	}

	correlationID := ""
	if vals[correlation.CorrelationIDKey] != nil {
		correlationID = vals[correlation.CorrelationIDKey].(string)
	} else {
		// If correlationId is missing, return the original logger
		return l
	}
	sublogger := l.With().Str(correlation.CorrelationIDKey, correlationID).Logger()

	return &Logger{Logger: &sublogger}
}

type LevelWriter struct {
	writer zerolog.LevelWriter
	level  zerolog.Level
}

func (w *LevelWriter) Write(p []byte) (n int, err error) {
	return w.writer.Write(p)
}
func (w *LevelWriter) WriteLevel(level zerolog.Level, p []byte) (n int, err error) {
	if level >= w.level {
		return w.writer.WriteLevel(level, p)
	}
	return len(p), nil
}

type Config struct {
	Path      string
	Prefix    string
	FileLevel string
	StdLevel  string
}

func GetLogger(cfg Config) (*Logger, error) {
	fileLevel, err := zerolog.ParseLevel(cfg.FileLevel)
	if err != nil {
		return nil, err
	}

	stdLevel, err := zerolog.ParseLevel(cfg.StdLevel)
	if err != nil {
		return nil, err
	}

	var logger zerolog.Logger
	if cfg.Path != "" {

		dirPath := filepath.Dir(cfg.Path)
		err := os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			log.Fatal("Could not create logger directory")
		}

		f, err := os.OpenFile(cfg.Path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			log.Println("error opening file", err)
			return nil, err
		}

		fileWriter := &LevelWriter{zerolog.MultiLevelWriter(f), fileLevel}
		stdWriter := &LevelWriter{zerolog.MultiLevelWriter(os.Stderr), stdLevel}

		multi := zerolog.MultiLevelWriter(fileWriter, stdWriter)
		logger = zerolog.New(multi).With().Timestamp().Logger()
	} else {
		logger = zerolog.New(os.Stderr).With().Timestamp().Logger().Level(stdLevel)
	}

	logger = logger.With().Caller().Logger()
	l := &Logger{&logger}

	return l, nil
}
