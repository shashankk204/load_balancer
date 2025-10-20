package logger
import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
)


type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	RequestID string `json:"request_id,omitempty"`
	Message   string `json:"message"`
	Method    string `json:"method,omitempty"`
	Path      string `json:"path,omitempty"`
	Target    string `json:"target,omitempty"`
	Duration  string `json:"duration,omitempty"`
	Status  string `json:"status,omitempty"`

}



type ctxKey string

const requestIDKey ctxKey = "request_id"


var (
	infoLogger  = log.New(os.Stdout, "", 0)
	errorLogger = log.New(os.Stderr, "", 0)
)


func WithRequestID(ctx context.Context) context.Context {
	return context.WithValue(ctx, requestIDKey, uuid.New().String())
}

func GetRequestID(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey).(string); ok {
		return v
	}
	return ""
}



func Info(ctx context.Context, msg string, fields map[string]string) {
	writeLog(ctx, "INFO", msg, fields, infoLogger)
}

func Error(ctx context.Context, msg string, fields map[string]string) {
	writeLog(ctx, "ERROR", msg, fields, errorLogger)
}

func writeLog(ctx context.Context, level, msg string, fields map[string]string, l *log.Logger) {
	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339Nano),
		Level:     level,
		RequestID: GetRequestID(ctx),
		Message:   msg,
	}

	if fields != nil {
		entry.Method = fields["method"]
		entry.Path = fields["path"]
		entry.Target = fields["target"]
		entry.Duration = fields["duration"]
		entry.Status= fields["status"]

	}

	data, _ := json.Marshal(entry)
	l.Println(string(data))
}