package log

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"path/filepath"

	"go.uber.org/zap/zapcore"
)

func init() {
	RegisterWriter(OutputConsole, DefaultConsoleWriterFactory)
	RegisterWriter(OutputFile, DefaultFileWriterFactory)
	DefaultLogger = NewZapLog(defaultConfig)
}

const pluginType = "log"

// default logger
var (
	DefaultConsoleWriterFactory = &ConsoleWriterFactory{}
	DefaultFileWriterFactory    = &FileWriterFactory{}

	writers = make(map[string]Factory)
	logs    = make(map[string]Logger)
)

// Register 注册日志，支持同时多个日志实现
func Register(name string, logger Logger) {
	logs[name] = logger
}

// RegisterWriter 注册日志输出writer，支持同时多个日志实现
func RegisterWriter(name string, writer Factory) {
	writers[name] = writer
}

// Get 通过日志名返回具体的实现 log.Debug使用DefaultLogger打日志，也可以使用 log.Get("name").Debug
func Get(name string) Logger {
	return logs[name]
}

// Factory 日志插件工厂 由框架启动读取配置文件 调用该工厂生成具体日志
type Factory interface {
	Setup(name string, configDec *Decoder) error
}

// Decoder log
type Decoder struct {
	OutputConfig *OutputConfig
	Core         zapcore.Core
}

// Decode 解析writer配置 复制一份
func (d *Decoder) Decode(conf interface{}) error {

	output, ok := conf.(**OutputConfig)
	if !ok {
		return fmt.Errorf("decoder config type:%T invalid, not **OutputConfig", conf)
	}

	*output = d.OutputConfig

	return nil
}

// Setup 启动加载log配置 并注册日志
func Setup(data []byte) error {
	var n2c map[string]Config
	err := yaml.Unmarshal(data, &n2c)
	if err != nil {
		return err
	}
	return SetupWithConfig(n2c)
}

func SetupWithConfig(n2c map[string]Config) error {
	for k, v := range n2c {
		// 如果没有配置caller skip，则默认为2
		callerSkip := 2
		for i := 0; i < len(v); i++ {
			if v[i].CallerSkip != 0 {
				callerSkip = v[i].CallerSkip
			}
		}

		logger := NewZapLogWithCallerSkip(v, callerSkip)

		Register(k, logger)

		if k == "default" {
			SetLogger(logger)
		}
	}

	return nil
}

// ConsoleWriterFactory  new console writer instance
type ConsoleWriterFactory struct {
}

// Setup 启动加载配置 并注册console output writer
func (f *ConsoleWriterFactory) Setup(name string, configDec *Decoder) error {
	if configDec == nil {
		return errors.New("console writer decoder empty")
	}
	decoder := configDec

	conf := &OutputConfig{}
	err := decoder.Decode(&conf)
	if err != nil {
		return err
	}

	decoder.Core = newConsoleCore(conf)
	return nil
}

// FileWriterFactory  new file writer instance
type FileWriterFactory struct {
}

// Setup 启动加载配置 并注册file output writer
func (f *FileWriterFactory) Setup(name string, configDec *Decoder) error {
	if configDec == nil {
		return errors.New("file writer decoder empty")
	}

	decoder := configDec

	conf := &OutputConfig{}
	err := decoder.Decode(&conf)
	if err != nil {
		return err
	}

	if conf.WriteConfig.LogPath != "" {
		conf.WriteConfig.Filename = filepath.Join(conf.WriteConfig.LogPath, conf.WriteConfig.Filename)
	}

	if conf.WriteConfig.RollType == "" {
		conf.WriteConfig.RollType = "size"
	}

	decoder.Core = newFileCore(conf)
	return nil
}
