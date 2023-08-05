package log

const (
	unknownKey = "unknown"
)

// TemporalLogger wraps the Logger struct and implements the
// Temporal Logger interface:
//
//     Logger interface {
//		Debug(msg string, keyvals ...interface{})
//		Info(msg string, keyvals ...interface{})
//		Warn(msg string, keyvals ...interface{})
//		Error(msg string, keyvals ...interface{})
//	  }
type TemporalLogger struct {
	logger *Logger
}

func NewTemporalLogger(logger *Logger) *TemporalLogger {
	return &TemporalLogger{logger: logger}
}

func (t *TemporalLogger) Debug(msg string, keyvals ...interface{}) {
	fields := t.fields(keyvals)
	t.logger.Debug(msg, fields...)
}

func (t *TemporalLogger) Info(msg string, keyvals ...interface{}) {
	fields := t.fields(keyvals)
	t.logger.Info(msg, fields...)
}

func (t *TemporalLogger) Warn(msg string, keyvals ...interface{}) {
	fields := t.fields(keyvals)
	t.logger.Warn(msg, fields...)
}

func (t *TemporalLogger) Error(msg string, keyvals ...interface{}) {
	fields := t.fields(keyvals)
	t.logger.Error(msg, fields...)
}

func (t *TemporalLogger) fields(keyvals ...interface{}) []Field {
	var fields []Field

	// TODO Temporal is adding the activity details to the logs after we log something in an activity.
	//  we should find a way to separate activity details from user logs.
	for i := range keyvals {
		field, ok := keyvals[i].(Field)
		if !ok {
			fields = append(fields, Any(unknownKey, keyvals[i]))
			continue
		}

		fields = append(fields, field)
	}

	return fields
}
