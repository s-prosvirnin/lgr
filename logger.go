package main

import (
	"context"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Debug(ctx context.Context, msg string, params ...KVParam)
	Info(ctx context.Context, msg string, params ...KVParam)
	Warn(ctx context.Context, err error, params ...KVParam)
	Error(ctx context.Context, err error, params ...KVParam)
	CtxWithParams(ctx context.Context, params ...KVParam) context.Context
}

type loggerConfig struct {
	Level    logLevel    `envconfig:"LEVEL" required:"true"`
	Encoding logEncoding `envconfig:"ENCODING" required:"true"`
}

type logLevel string

const (
	logLevelDebug   logLevel = "debug"
	logLevelInfo    logLevel = "info"
	logLevelWarning logLevel = "warning"
	logLevelError   logLevel = "error"
)

const contextParamsKey = "log_params"

func (l logLevel) toZapLevel() (zap.AtomicLevel, error) {
	switch l {
	case logLevelDebug:
		return zap.NewAtomicLevelAt(zapcore.DebugLevel), nil
	case logLevelError:
		return zap.NewAtomicLevelAt(zapcore.ErrorLevel), nil
	case logLevelInfo:
		return zap.NewAtomicLevelAt(zapcore.InfoLevel), nil
	case logLevelWarning:
		return zap.NewAtomicLevelAt(zapcore.WarnLevel), nil
	default:
		return zap.AtomicLevel{}, errors.Errorf("unknown log level: %v", l)
	}
}

type logEncoding string

const (
	logTypeText logEncoding = "text"
	logTypeJson logEncoding = "json"
)

func (t logEncoding) toZapEncoding() (string, error) {
	switch t {
	case logTypeText:
		return "console", nil
	case logTypeJson:
		return "json", nil
	default:
		return "", errors.Errorf("unknown log encoding: %v", t)
	}
}

type logger struct {
	zLogger *zap.Logger
}

var _ Logger = (*logger)(nil)

func newLogger(zLogger *zap.Logger) logger {
	return logger{
		zLogger: zLogger,
	}
}

func (l logger) Error(ctx context.Context, err error, params ...KVParam) {
	var tErr Error
	if errors.As(err, &tErr) {
		params = l.mergeParams(
			params,
			tErr.Params(),
			getContextParams(ctx),
			[]KVParam{KVString("err_type", tErr.Type().String())},
		)
	} else {
		params = l.mergeParams(
			params, getContextParams(ctx),
		)
	}
	l.zLogger.Error(err.Error(), l.convertToZapFields(params)...)
}

func (l logger) Warn(ctx context.Context, err error, params ...KVParam) {
	var tErr Error
	if errors.As(err, &tErr) {
		params = l.mergeParams(
			params,
			tErr.Params(),
			getContextParams(ctx),
			[]KVParam{KVString("err_type", tErr.Type().String())},
		)
	} else {
		params = l.mergeParams(
			params, getContextParams(ctx),
		)
	}
	l.zLogger.Warn(err.Error(), l.convertToZapFields(params)...)
}

func (l logger) Info(ctx context.Context, msg string, params ...KVParam) {
	l.zLogger.Info(msg, l.convertToZapFields(l.mergeParams(params, getContextParams(ctx)))...)
}

func (l logger) Debug(ctx context.Context, msg string, params ...KVParam) {
	l.zLogger.Debug(msg, l.convertToZapFields(l.mergeParams(params, getContextParams(ctx)))...)
}

func (l logger) CtxWithParams(ctx context.Context, params ...KVParam) context.Context {
	return CtxWithValue(ctx, params...)
}

func (l logger) Flush() {
	l.zLogger.Sync()
}

func (l logger) mergeParams(paramsList ...[]KVParam) []KVParam {
	capacity := 0
	for _, params := range paramsList {
		capacity += len(params)
	}
	newParams := make([]KVParam, 0, capacity)
	for _, params := range paramsList {
		newParams = append(newParams, params...)
	}

	return newParams
}

func (l logger) convertToZapFields(params []KVParam) []zap.Field {
	zapFields := make([]zap.Field, 0, len(params))
	// Map for parameter deduplication.
	uniqueParams := make(map[string]struct{}, len(params))
	for _, param := range params {
		if _, ok := uniqueParams[param.Key()]; ok {
			continue
		}
		uniqueParams[param.Key()] = struct{}{}
		zapFields = append(zapFields, l.convertToZapField(param))
	}

	return zapFields
}

func (l logger) convertToZapField(param KVParam) zap.Field {
	switch param.Type() {
	case KVParamTypeString:
		return zap.String(param.Key(), param.String())
	case KVParamTypeInt:
		return zap.Int(param.Key(), param.Int())
	case KVParamTypeFloat:
		return zap.Float64(param.Key(), param.Float())
	case KVParamTypeBoolean:
		return zap.Bool(param.Key(), param.Bool())
	case KVParamTypeTime:
		return zap.Time(param.Key(), param.Time())
	case KVParamTypeDuration:
		return zap.Duration(param.Key(), param.Duration())
	default:
		return zap.Any(param.Key(), "")
	}
}

func initLogger(cfg loggerConfig) (logger, error) {
	zapLevel, err := cfg.Level.toZapLevel()
	if err != nil {
		return logger{}, err
	}
	zapEncoding, err := cfg.Encoding.toZapEncoding()
	if err != nil {
		return logger{}, err
	}

	zapConfig := zap.NewProductionConfig()
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zapConfig.Encoding = zapEncoding
	zapConfig.Sampling = nil           // collapsing records with the same message and level (for some reason, parameters are not taken into account)
	zapConfig.DisableStacktrace = true // do not log stacktrace
	zapConfig.Level = zapLevel
	zapConfig.OutputPaths = []string{"stdout"}
	zapConfig.ErrorOutputPaths = []string{"stdout"}

	zapLogger, err := zapConfig.Build(
		zap.AddCaller(), zap.AddCallerSkip(1),
	)
	if err != nil {
		return logger{}, errors.Wrap(err, "build zap logger error")
	}

	lgr := newLogger(zapLogger)

	return lgr, nil
}
