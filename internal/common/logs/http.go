package logs

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
)

func NewStructuredLogger(logger *logrus.Logger) func(next http.Handler) http.Handler {
	return middleware.RequestLogger(&StructuredLogger{
		Logger: logger,
	})
}

// based on example from chi: https://github.com/go-chi/chi/blob/master/_examples/logging/main.go
type StructuredLogger struct {
	Logger *logrus.Logger
}

func (logger *StructuredLogger) NewLogEntry(r *http.Request) middleware.LogEntry {
	entry := &StructuredLoggerEntry{
		Logger: logrus.NewEntry(logger.Logger),
	}
	logFields := logrus.Fields{}

	if reqID := middleware.GetReqID(r.Context()); reqID != "" {
		logFields["req_id"] = reqID
	}

	logFields["http_method"] = r.Method

	logFields["remote_addr"] = r.RemoteAddr
	logFields["uri"] = r.RequestURI

	entry.Logger = entry.Logger.WithFields(logFields)

	entry.Logger.Info("Request started")

	return entry

}

type StructuredLoggerEntry struct {
	Logger logrus.FieldLogger
}

func (loggerEntry *StructuredLoggerEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	loggerEntry.Logger = loggerEntry.Logger.WithFields(logrus.Fields{
		"resp_status":       status,
		"resp_bytes_length": bytes,
		"resp_elapsed":      elapsed.Round(time.Millisecond / 100).String(),
	})

	loggerEntry.Logger.Info("Request completed	")
}

func (loggerEntry *StructuredLoggerEntry) Panic(v interface{}, stack []byte) {
	loggerEntry.Logger = loggerEntry.Logger.WithFields(logrus.Fields{
		"stack": string(stack),
		"panic": fmt.Sprintf("%+v", v),
	})
}

func GetLogEntry(r *http.Request) logrus.FieldLogger {
	entry := middleware.GetLogEntry(r).(*StructuredLoggerEntry)
	return entry.Logger
}
