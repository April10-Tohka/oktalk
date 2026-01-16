package log

import (
	"oktalk/internal/pkg/constants"

	"github.com/sirupsen/logrus"
)

type TraceContextHook struct{}

func (h *TraceContextHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *TraceContextHook) Fire(entry *logrus.Entry) error {
	if entry.Context != nil {
		// 从 context 中获取 TraceID
		if traceID, ok := entry.Context.Value(constants.TraceIDKey).(string); ok {
			entry.Data[constants.TraceIDKey] = traceID
		}
	}
	return nil
}
