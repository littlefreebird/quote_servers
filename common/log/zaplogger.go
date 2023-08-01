package log

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

var defaultConfig = []OutputConfig{
	{
		Writer:    "console",
		Level:     "debug",
		Formatter: "console",
	},
}

// core常量定义
const (
	ConsoleZapCore = "console"
	FileZapCore    = "file"
)

var atomicLevels = make(map[string]zap.AtomicLevel)

// Levels zapcore level
var Levels = map[string]zapcore.Level{
	"":      zapcore.DebugLevel,
	"debug": zapcore.DebugLevel,
	"info":  zapcore.InfoLevel,
	"warn":  zapcore.WarnLevel,
	"error": zapcore.ErrorLevel,
	"fatal": zapcore.FatalLevel,
}

var levelToZapLevel = map[Level]zapcore.Level{
	LevelTrace: zapcore.DebugLevel,
	LevelDebug: zapcore.DebugLevel,
	LevelInfo:  zapcore.InfoLevel,
	LevelWarn:  zapcore.WarnLevel,
	LevelError: zapcore.ErrorLevel,
	LevelFatal: zapcore.FatalLevel,
}

var zapLevelToLevel = map[zapcore.Level]Level{
	zapcore.DebugLevel: LevelDebug,
	zapcore.InfoLevel:  LevelInfo,
	zapcore.WarnLevel:  LevelWarn,
	zapcore.ErrorLevel: LevelError,
	zapcore.FatalLevel: LevelFatal,
}

// NewZapLog 创建一个默认实现的logger, callerskip为2
func NewZapLog(c Config) Logger {
	return NewZapLogWithCallerSkip(c, 2)
}

// NewZapLogWithCallerSkip 创建一个默认实现的logger
func NewZapLogWithCallerSkip(c Config, callerSkip int) Logger {

	cores := make([]zapcore.Core, 0, len(c))
	for _, o := range c {

		writer, ok := writers[o.Writer]
		if !ok {
			fmt.Printf("log writer core:%s no registered!\n", o.Writer)
			continue
		}

		decoder := &Decoder{OutputConfig: &o}
		err := writer.Setup(o.Writer, decoder)
		if err != nil {
			fmt.Printf("log writer setup core:%s fail:%v!\n", o.Writer, err)
			continue
		}

		cores = append(cores, decoder.Core)
	}

	return &zapLog{
		logger: zap.New(zapcore.NewTee(cores...),
			zap.AddCallerSkip(callerSkip),
			zap.AddCaller()),
	}
}

func newEncoder(c *OutputConfig) zapcore.Encoder {
	encoderCfg := zapcore.EncoderConfig{
		TimeKey:        "T",
		LevelKey:       "L",
		NameKey:        "N",
		CallerKey:      "C",
		MessageKey:     "M",
		StacktraceKey:  "S",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     NewTimeEncoder(c.FormatConfig.TimeFmt),
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	encoder := zapcore.NewConsoleEncoder(encoderCfg)

	switch c.Formatter {
	case "console":
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	case "json":
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	}
	return encoder
}

func newConsoleCore(c *OutputConfig) zapcore.Core {
	lvl := zap.NewAtomicLevelAt(Levels[c.Level])
	atomicLevels[OutputConsole] = lvl
	return zapcore.NewCore(
		newEncoder(c),
		zapcore.Lock(os.Stdout),
		lvl)
}

func newFileCore(c *OutputConfig) zapcore.Core {
	var ws zapcore.WriteSyncer

	if c.WriteConfig.RollType == "size" {
		// 按大小滚动
		ws = zapcore.AddSync(&lumberjack.Logger{
			Filename:   c.WriteConfig.Filename,
			MaxSize:    c.WriteConfig.MaxSize,
			MaxBackups: c.WriteConfig.MaxBackups,
			MaxAge:     c.WriteConfig.MaxAge,
			Compress:   c.WriteConfig.Compress,
		})
	} else {
		// 按时间滚动
		hook, _ := rotatelogs.New(
			c.WriteConfig.Filename+c.WriteConfig.TimeUnit.Format(),
			rotatelogs.WithMaxAge(time.Duration(int64(24*time.Hour)*int64(c.WriteConfig.MaxAge))),
			rotatelogs.WithRotationTime(c.WriteConfig.TimeUnit.RotationGap()),
		)
		ws = zapcore.AddSync(hook)
	}

	lvl := zap.NewAtomicLevelAt(Levels[c.Level])
	atomicLevels[OutputFile] = lvl

	return zapcore.NewCore(
		newEncoder(c),
		ws, lvl,
	)
}

// NewTimeEncoder 创建时间格式encoder
func NewTimeEncoder(format string) zapcore.TimeEncoder {
	if format == "" {
		return func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendByteString(DefaultTimeFormat(t))
		}
	}
	return func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(CustomTimeFormat(t, format))
	}
}

// CustomTimeFormat 自定义时间格式
func CustomTimeFormat(t time.Time, format string) string {
	return t.Format(format)
}

// DefaultTimeFormat 默认时间格式
func DefaultTimeFormat(t time.Time) []byte {
	t = t.Local()
	year, month, day := t.Date()
	hour, minute, second := t.Clock()
	micros := t.Nanosecond() / 1000

	buf := make([]byte, 23)
	buf[0] = byte((year/1000)%10) + '0'
	buf[1] = byte((year/100)%10) + '0'
	buf[2] = byte((year/10)%10) + '0'
	buf[3] = byte(year%10) + '0'
	buf[4] = '-'
	buf[5] = byte((month)/10) + '0'
	buf[6] = byte((month)%10) + '0'
	buf[7] = '-'
	buf[8] = byte((day)/10) + '0'
	buf[9] = byte((day)%10) + '0'
	buf[10] = ' '
	buf[11] = byte((hour)/10) + '0'
	buf[12] = byte((hour)%10) + '0'
	buf[13] = ':'
	buf[14] = byte((minute)/10) + '0'
	buf[15] = byte((minute)%10) + '0'
	buf[16] = ':'
	buf[17] = byte((second)/10) + '0'
	buf[18] = byte((second)%10) + '0'
	buf[19] = '.'
	buf[20] = byte((micros/100000)%10) + '0'
	buf[21] = byte((micros/10000)%10) + '0'
	buf[22] = byte((micros/1000)%10) + '0'
	return buf
}

// zapLogWrapper 是对 zapLogger 的代理，引入 zapLogWrapper 的原因见
// 通过引入 zapLogWrapper 这个代理，使 debug 系列函数的调用增加一层，让 caller 信息能够正确的设置
type zapLogWrapper struct {
	l *zapLog
}

// Trace logs to TRACE log, Arguments are handled in the manner of fmt.Print
func (z *zapLogWrapper) Trace(args ...interface{}) {
	z.l.Trace(args...)
}

// Tracef logs to TRACE log, Arguments are handled in the manner of fmt.Printf
func (z *zapLogWrapper) Tracef(format string, args ...interface{}) {
	z.l.Tracef(format, args...)
}

// Debug logs to DEBUG log, Arguments are handled in the manner of fmt.Print
func (z *zapLogWrapper) Debug(args ...interface{}) {
	z.l.Debug(args...)
}

// Debugf logs to DEBUG log, Arguments are handled in the manner of fmt.Printf
func (z *zapLogWrapper) Debugf(format string, args ...interface{}) {
	z.l.Debugf(format, args...)
}

// Info logs to INFO log, Arguments are handled in the manner of fmt.Print
func (z *zapLogWrapper) Info(args ...interface{}) {
	z.l.Info(args...)
}

// Infof logs to INFO log, Arguments are handled in the manner of fmt.Printf
func (z *zapLogWrapper) Infof(format string, args ...interface{}) {
	z.l.Infof(format, args...)
}

// Warn logs to WARNING log, Arguments are handled in the manner of fmt.Print
func (z *zapLogWrapper) Warn(args ...interface{}) {
	z.l.Warn(args...)
}

// Warnf logs to WARNING log, Arguments are handled in the manner of fmt.Printf
func (z *zapLogWrapper) Warnf(format string, args ...interface{}) {
	z.l.Warnf(format, args...)
}

// Error logs to ERROR log, Arguments are handled in the manner of fmt.Print
func (z *zapLogWrapper) Error(args ...interface{}) {
	z.l.Error(args...)
}

// Errorf logs to ERROR log, Arguments are handled in the manner of fmt.Printf
func (z *zapLogWrapper) Errorf(format string, args ...interface{}) {
	z.l.Errorf(format, args...)
}

// Fatal logs to FATAL log, Arguments are handled in the manner of fmt.Print
func (z *zapLogWrapper) Fatal(args ...interface{}) {
	z.l.Fatal(args...)
}

// Fatalf logs to FATAL log, Arguments are handled in the manner of fmt.Printf
func (z *zapLogWrapper) Fatalf(format string, args ...interface{}) {
	z.l.Fatalf(format, args...)
}

// SetLevel 设置输出端日志级别
func (z *zapLogWrapper) SetLevel(output string, level Level) {
	z.l.SetLevel(output, level)
}

// GetLevel 获取输出端日志级别
func (z *zapLogWrapper) GetLevel(output string) Level {
	return z.l.GetLevel(output)
}

// WithFields 设置一些业务自定义数据到每条log里:比如uid，imei等, 每个请求入口设置，并生成一个新的logger，后续使用新的logger来打日志 fields 必须kv成对出现
func (z *zapLogWrapper) WithFields(fields ...string) Logger {
	return z.l.WithFields(fields...)
}

type zapLog struct {
	logger *zap.Logger
}

// WithFields 设置一些业务自定义数据到每条log里:比如uid，imei等, 每个请求入口设置，并生成一个新的logger，后续使用新的logger来打日志 fields 必须kv成对出现
func (l *zapLog) WithFields(fields ...string) Logger {

	zapfields := make([]zap.Field, len(fields)/2)
	for index := range zapfields {
		zapfields[index] = zap.String(fields[2*index], fields[2*index+1])
	}

	// 使用 zapLogWrapper 代理，这样返回的 Logger 被调用时，调用栈层数和使用 Debug 系列函数一致，caller 信息能够正确的设置
	return &zapLogWrapper{l: &zapLog{logger: l.logger.With(zapfields...)}}
}

// Trace logs to TRACE log, Arguments are handled in the manner of fmt.Print
func (l *zapLog) Trace(args ...interface{}) {
	if l.logger.Core().Enabled(zapcore.DebugLevel) {
		l.logger.Debug(fmt.Sprint(args...))
	}
}

// Tracef logs to TRACE log, Arguments are handled in the manner of fmt.Printf
func (l *zapLog) Tracef(format string, args ...interface{}) {
	if l.logger.Core().Enabled(zapcore.DebugLevel) {
		l.logger.Debug(fmt.Sprintf(format, args...))
	}
}

// Debug logs to DEBUG log, Arguments are handled in the manner of fmt.Print
func (l *zapLog) Debug(args ...interface{}) {
	if l.logger.Core().Enabled(zapcore.DebugLevel) {
		l.logger.Debug(fmt.Sprint(args...))
	}
}

// Debugf logs to DEBUG log, Arguments are handled in the manner of fmt.Printf
func (l *zapLog) Debugf(format string, args ...interface{}) {
	if l.logger.Core().Enabled(zapcore.DebugLevel) {
		l.logger.Debug(fmt.Sprintf(format, args...))
	}
}

// Info logs to INFO log, Arguments are handled in the manner of fmt.Print
func (l *zapLog) Info(args ...interface{}) {
	if l.logger.Core().Enabled(zapcore.InfoLevel) {
		l.logger.Info(fmt.Sprint(args...))
	}
}

// Infof logs to INFO log, Arguments are handled in the manner of fmt.Printf
func (l *zapLog) Infof(format string, args ...interface{}) {
	if l.logger.Core().Enabled(zapcore.InfoLevel) {
		l.logger.Info(fmt.Sprintf(format, args...))
	}
}

// Warn logs to WARNING log, Arguments are handled in the manner of fmt.Print
func (l *zapLog) Warn(args ...interface{}) {
	if l.logger.Core().Enabled(zapcore.WarnLevel) {
		l.logger.Warn(fmt.Sprint(args...))
	}
}

// Warnf logs to WARNING log, Arguments are handled in the manner of fmt.Printf
func (l *zapLog) Warnf(format string, args ...interface{}) {
	if l.logger.Core().Enabled(zapcore.WarnLevel) {
		l.logger.Warn(fmt.Sprintf(format, args...))
	}
}

// Error logs to ERROR log, Arguments are handled in the manner of fmt.Print
func (l *zapLog) Error(args ...interface{}) {
	if l.logger.Core().Enabled(zapcore.ErrorLevel) {
		l.logger.Error(fmt.Sprint(args...))
	}
}

// Errorf logs to ERROR log, Arguments are handled in the manner of fmt.Printf
func (l *zapLog) Errorf(format string, args ...interface{}) {
	if l.logger.Core().Enabled(zapcore.ErrorLevel) {
		l.logger.Error(fmt.Sprintf(format, args...))
	}
}

// Fatal logs to FATAL log, Arguments are handled in the manner of fmt.Print
func (l *zapLog) Fatal(args ...interface{}) {
	if l.logger.Core().Enabled(zapcore.FatalLevel) {
		l.logger.Fatal(fmt.Sprint(args...))
	}
}

// Fatalf logs to FATAL log, Arguments are handled in the manner of fmt.Printf
func (l *zapLog) Fatalf(format string, args ...interface{}) {
	if l.logger.Core().Enabled(zapcore.FatalLevel) {
		l.logger.Fatal(fmt.Sprintf(format, args...))
	}
}

// SetLevel 设置输出端日志级别
func (l *zapLog) SetLevel(output string, level Level) {
	if atomLevel, ok := atomicLevels[output]; ok {
		atomLevel.SetLevel(levelToZapLevel[level])
	}
}

// GetLevel 获取输出端日志级别
func (l *zapLog) GetLevel(output string) Level {
	if atomLevel, ok := atomicLevels[output]; ok {
		return zapLevelToLevel[atomLevel.Level()]
	}
	return LevelDebug
}
