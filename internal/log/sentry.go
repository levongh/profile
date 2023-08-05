package log

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"go.uber.org/zap/zapcore"
)

type sentryOption struct {
	sentryDsn    string
	sentryTags   map[string]string
	sentryFields []zapcore.Field
}

func WithSentry(dsn string, tags map[string]string, fields ...zapcore.Field) Option {
	return func(logger *Logger) {
		logger.sentryOption.sentryDsn = dsn
		logger.sentryOption.sentryTags = tags
		logger.sentryOption.sentryFields = fields
	}
}

const (
	serviceTag = "service"

	// sentry specific
	prefixTagZapField = "prefixTagZapField"
	zapFieldPrefix    = "zapfield_prefix"
	errorFieldName    = "_error"
	errorKeyName      = "error"
)

type SentryCore struct {
	flushTimeout time.Duration
	client       Sentryer
	level        zapcore.Level
	tags         map[string]string
	fields       []zapcore.Field
}

// SentryOptions advanced options for sentry client
type SentryOptions struct {
	sentry.ClientOptions
	MinLevel     zapcore.Level
	FlushTimeout time.Duration
	Tags         map[string]string
}

type Sentryer interface {
	Flush(timeout time.Duration) bool
	Recover(err interface{}, hint *sentry.EventHint, scope sentry.EventModifier) *sentry.EventID
	CaptureException(exception error, hint *sentry.EventHint, scope sentry.EventModifier) *sentry.EventID
	CaptureMessage(message string, hint *sentry.EventHint, scope sentry.EventModifier) *sentry.EventID
}

// newSentryOptions returns sentry options
func newSentryOptions(dsn, mode, serviceName string) SentryOptions {
	options := SentryOptions{
		ClientOptions: sentry.ClientOptions{
			Dsn:         dsn,
			Debug:       mode == modeDev,
			Environment: mode,
		},
		FlushTimeout: 0,
		Tags:         make(map[string]string, 1),
	}
	options.Tags[serviceTag] = serviceName
	return options
}

// newSentryCore returns sentry core
func newSentryCore(options SentryOptions) (*SentryCore, error) {
	client, err := sentry.NewClient(options.ClientOptions)
	if err != nil {
		return nil, err
	}
	timeout := options.FlushTimeout
	if timeout.Seconds() <= 3 {
		timeout = time.Second * 3
	}
	newCore := &SentryCore{
		flushTimeout: timeout,
		client:       client,
		level:        options.MinLevel,
		tags:         make(map[string]string),
		fields:       make([]zapcore.Field, 0),
	}
	for key, value := range options.Tags {
		newCore.tags[key] = value
	}
	return newCore, nil
}

func (s *SentryCore) Check(entry zapcore.Entry, check *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if s.Enabled(entry.Level) {
		return check.AddCore(entry, s)
	}
	return check
}

func (s *SentryCore) Enabled(lvl zapcore.Level) bool {
	return s.level <= lvl
}

func (s *SentryCore) Sync() error {
	if !s.client.Flush(s.flushTimeout) {
		return errors.New("flush failed")
	}
	return nil
}

func (s *SentryCore) With(fields []zapcore.Field) zapcore.Core {
	return s.clone(fields)
}

func (s *SentryCore) clone(fields []zapcore.Field) *SentryCore {
	cloned := &SentryCore{
		client: s.client,
		level:  s.level,
		tags:   make(map[string]string),
		fields: make([]zapcore.Field, 0, len(s.fields)),
	}

	// Clone tags
	for key, value := range s.tags {
		cloned.tags[key] = value
	}

	// Clone fields
	var arrayFields []zapcore.Field
	copy(arrayFields, s.fields)
	arrayFields = append(arrayFields, fields...)
	cloned.fields = append(cloned.fields, arrayFields...)

	return cloned
}

func findField(name string, fields []zapcore.Field) (zapcore.Field, bool) {
	for _, field := range fields {
		if field.Key == name {
			return field, true
		}
	}
	return zapcore.Field{}, false
}

func getErrorFromFields(entry zapcore.Entry, fields []zapcore.Field) error {
	if errField, ok := findField(errorFieldName, fields); ok && errField.Type == zapcore.ErrorType {
		return errField.Interface.(error)
	}
	return errors.New(entry.Message)
}

func (s *SentryCore) applyFieldsToScope(scope *sentry.Scope, _ zapcore.Entry, fields []zapcore.Field) error {
	extras := make(map[string]interface{}, len(fields))
	var arrayFields []zapcore.Field
	copy(arrayFields, fields)
	arrayFields = append(arrayFields, s.fields...)
	for _, field := range arrayFields {
		if !strings.HasPrefix(field.Key, "_") &&
			strings.ToLower(field.Key) != errorKeyName &&
			!strings.HasPrefix(field.Key, zapFieldPrefix) {
			value, err := fieldValueAsInterface(field)
			if err != nil {
				return err
			}
			extras[field.Key] = value
		}
	}
	scope.SetExtras(extras)
	return nil
}

func (s *SentryCore) applyTagsToScope(scope *sentry.Scope, fields []zapcore.Field) error {
	scope.SetTags(s.tags)
	var arrayFields []zapcore.Field
	copy(arrayFields, fields)
	arrayFields = append(arrayFields, s.fields...)
	for _, field := range arrayFields {
		if strings.HasPrefix(field.Key, prefixTagZapField) {
			tag := field.Key[4:]
			value, err := fieldValueAsInterface(field)
			if err != nil {
				return err
			}
			if tag != "" {
				valueStr := fmt.Sprintf("%v", value)
				scope.SetTag(tag, valueStr)
			}
		}
	}
	return nil
}

func (s *SentryCore) createScope(entry zapcore.Entry, fields []zapcore.Field) (*sentry.Scope, error) {
	scope := sentry.NewScope()
	levelSentry := levelToSentryLevel(entry.Level)
	scope.SetLevel(levelSentry)
	if err := s.applyFieldsToScope(scope, entry, fields); err != nil {
		return nil, err
	}
	if err := s.applyTagsToScope(scope, fields); err != nil {
		return nil, err
	}
	return scope, nil
}

func (s *SentryCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	scope, err := s.createScope(entry, fields)
	if err != nil {
		return err
	}

	switch entry.Level {
	case zapcore.ErrorLevel:
		err := getErrorFromFields(entry, fields)
		if _, ok := err.(*PanicWrapper); ok {
			scope.SetLevel(sentry.LevelFatal)
			s.client.Recover(err, &sentry.EventHint{RecoveredException: err}, scope)
		} else {
			s.client.CaptureException(err,
				&sentry.EventHint{Data: entry.Message, OriginalException: err}, scope)
		}
	case zapcore.PanicLevel, zapcore.DPanicLevel:
		err := getErrorFromFields(entry, fields)
		s.client.Recover(err, &sentry.EventHint{RecoveredException: err}, scope)
	case zapcore.InfoLevel, zapcore.WarnLevel, zapcore.DebugLevel:
		// TODO: do we need to post to sentry info/warn/debug levels?
		s.client.CaptureMessage(entry.Message, nil, scope)
	default:
		return fmt.Errorf("%s: %s", errors.New("unknown level"), entry.Level.String())
	}
	return nil
}

func levelToSentryLevel(level zapcore.Level) sentry.Level {
	switch level {
	case zapcore.DebugLevel:
		return sentry.LevelDebug
	case zapcore.InfoLevel:
		return sentry.LevelInfo
	case zapcore.WarnLevel:
		return sentry.LevelWarning
	case zapcore.ErrorLevel:
		return sentry.LevelError
	case zapcore.DPanicLevel:
		return sentry.LevelFatal
	case zapcore.PanicLevel:
		return sentry.LevelFatal
	case zapcore.FatalLevel:
		return sentry.LevelFatal
	default:
		panic(fmt.Errorf("unknown level: %v", level))
	}
}

func fieldValueAsInterface(field zapcore.Field) (interface{}, error) {
	switch field.Type {
	case zapcore.BinaryType:
		return field.Interface.([]byte), nil
	case zapcore.BoolType:
		return field.Integer == 1, nil
	case zapcore.ByteStringType:
		return field.Interface.([]byte), nil
	case zapcore.Complex128Type:
		return field.Interface.(complex128), nil
	case zapcore.Complex64Type:
		return field.Interface.(complex64), nil
	case zapcore.DurationType:
		return time.Duration(field.Integer), nil
	case zapcore.Float64Type:
		return math.Float64frombits(uint64(field.Integer)), nil
	case zapcore.Float32Type:
		return math.Float32frombits(uint32(field.Integer)), nil
	case zapcore.Int64Type:
		return field.Integer, nil
	case zapcore.Int32Type:
		return int32(field.Integer), nil
	case zapcore.Int16Type:
		return int16(field.Integer), nil
	case zapcore.Int8Type:
		return int8(field.Integer), nil
	case zapcore.StringType:
		return field.String, nil
	case zapcore.TimeType:
		if field.Interface != nil {
			return time.Unix(0, field.Integer).In(field.Interface.(*time.Location)), nil
		}
		// Fall back to UTC if location is nil.
		return time.Unix(0, field.Integer), nil

	case zapcore.Uint64Type:
		return uint64(field.Integer), nil
	case zapcore.Uint32Type:
		return uint32(field.Integer), nil
	case zapcore.Uint16Type:
		return uint16(field.Integer), nil
	case zapcore.Uint8Type:
		return uint8(field.Integer), nil
	case zapcore.UintptrType:
		return uintptr(field.Integer), nil
	case zapcore.StringerType:
		return field.Interface.(fmt.Stringer).String(), nil
	case zapcore.ErrorType:
		return field.Interface.(error), nil
	case zapcore.SkipType:
		break
	}
	return nil, fmt.Errorf("unknown field type: %v", field.Type)
}

type PanicWrapper struct {
	error
}
