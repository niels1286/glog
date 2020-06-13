// @Title
// @Description
// @Author  Niels  2020/4/30
package glog

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strings"
	"time"
)

var LoggersMap = map[string]*Logger{}

type LoggerFactory struct {
	LoggerMap map[string]*Logger
}

type Logger struct {
	Name   string
	Level  string
	writer io.Writer
	Debug  *log.Logger // 记录所有日志
	Info   *log.Logger // 重要的信息
	Warn   *log.Logger // 需要注意的信息
	Error  *log.Logger // 非常严重的问题
}

var Debug *log.Logger
var Info *log.Logger
var Warn *log.Logger
var Error *log.Logger

func init() {
	cfg := ReadLogCfg()
	if cfg.Autoload {
		//自动加载配置文件
		ticker := time.NewTicker(time.Second * 300)
		go func() {
			for _ = range ticker.C {
				reloadCfgAndChangeLoggers(cfg)
			}
		}()
	}
	cfgs := []*LoggerCfg{}
	for key, val := range cfg.LoggerCfgs {

		if !strings.HasSuffix(val.File, ".log") {
			val.File = val.File + ".log"
		}
		if val.MaxBackupIndex < 1 {
			val.MaxBackupIndex = 1
		}

		logger := &Logger{
			Name:   key,
			Level:  val.Level,
			writer: getWriter(cfg.Root, val.File, val.Console),
		}
		logger.initLogger(val)
		LoggersMap[key] = logger
		cfgs = append(cfgs, val)
	}
	Debug = GetLogger("").Debug
	Info = GetLogger("").Info
	Warn = GetLogger("").Warn
	Error = GetLogger("").Error
	//在这里启动rolling
	restartTask(cfg)
}

func (l *Logger) initLogger(cfg *LoggerCfg) {
	l.Debug = log.New(l.writer,
		l.getPrefix()+"DEBUG: ",
		log.Ldate|log.Ltime|log.Llongfile)

	l.Info = log.New(l.writer,
		l.getPrefix()+"INFO : ",
		log.Ldate|log.Ltime|log.Llongfile)

	l.Warn = log.New(l.writer,
		l.getPrefix()+"WARN : ",
		log.Ldate|log.Ltime|log.Llongfile)

	l.Error = log.New(l.writer,
		l.getPrefix()+"ERROR: ",
		log.Ldate|log.Ltime|log.Llongfile)

	switch strings.ToLower(l.Level) {
	case "debug":
	case "info":
		l.Debug.SetOutput(ioutil.Discard)
	case "warn":
		l.Debug.SetOutput(ioutil.Discard)
		l.Info.SetOutput(ioutil.Discard)
	case "error":
		l.Debug.SetOutput(ioutil.Discard)
		l.Info.SetOutput(ioutil.Discard)
		l.Warn.SetOutput(ioutil.Discard)
	}
}

func (l *Logger) getPrefix() string {
	if l.Name == "default" {
		return ""
	}
	return strings.ToUpper(l.Name) + "-"
}

func (l *Logger) refreshWriter(root string, cfg *LoggerCfg) {
	l.writer = getWriter(root, cfg.File, cfg.Console)
	switch strings.ToLower(l.Level) {
	case "debug":
		l.Debug.SetOutput(l.writer)
		l.Info.SetOutput(l.writer)
		l.Warn.SetOutput(l.writer)
		l.Error.SetOutput(l.writer)
	case "info":
		l.Info.SetOutput(l.writer)
		l.Warn.SetOutput(l.writer)
		l.Error.SetOutput(l.writer)
	case "warn":
		l.Warn.SetOutput(l.writer)
		l.Error.SetOutput(l.writer)
	case "error":
		l.Error.SetOutput(l.writer)
	}
}

func getWriter(root string, filePath string, console bool) io.Writer {
	os.Mkdir(root, os.ModePerm)
	file, err := os.OpenFile(root+filePath,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open error log file:", err)
	}

	if console {
		return io.MultiWriter(file, os.Stdout)
	}

	return io.MultiWriter(file)
}

func reloadCfgAndChangeLoggers(cfg *LogCfg) {
	newCfg := ReadLogCfg()
	if reflect.DeepEqual(newCfg, cfg) {
		return
	}
	newLoggersMap := map[string]*Logger{}
	for key, val := range newCfg.LoggerCfgs {
		logger := &Logger{
			Name:   key,
			Level:  val.Level,
			writer: getWriter(newCfg.Root, val.File, val.Console),
		}
		logger.initLogger(val)
		newLoggersMap[key] = logger
	}
	LoggersMap = newLoggersMap
}

func GetLogger(name string) *Logger {
	if name == "" {
		name = "default"
	}
	return LoggersMap[name]
}
