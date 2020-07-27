// @Title
// @Description
// @Author  Niels  2020/4/30
package glog

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

type LoggerCfg struct {
	Level          string `yaml:"level"`
	File           string `yaml:"file"`
	Compress       bool   `yaml:"compress"`
	MaxFileSize    string `yaml:"maxFileSize"`
	MaxBackupIndex int    `yaml:"maxBackupIndex"`
	Console        bool   `yaml:"console"`
}

type LogCfg struct {
	Root       string                `yaml:"root"`
	Autoload   bool                  `yaml:"autoload"`
	LoggerCfgs map[string]*LoggerCfg `yaml:"loggerCfgs"`
}

func ReadLogCfg() *LogCfg {
	yamlFile, err := ioutil.ReadFile("log.yaml")
	if err != nil {
		log.Printf("log.yaml.Get warn   #%v ", err)
		yamlFile = []byte("root: 'logs/'\nautoload: true\nloggerCfgs:\n  default:\n    level: 'info'\n    file: 'def.log'\n    compress: true\n    maxFileSize: 100MB\n    maxBackupIndex: 7\n    console: true\n")
	}
	cfg := &LogCfg{}
	err = yaml.Unmarshal(yamlFile, cfg)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	return cfg
}
