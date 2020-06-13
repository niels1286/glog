// @Title
// @Description
// @Author  Niels  2020/5/27
package glog

import (
	"compress/gzip"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

var cfg *LogCfg
var running bool

func restartTask(newCfg *LogCfg) {
	cfg = newCfg
	if running {
		return
	}
	//date按天、week按周、month按月、none不生成新的文件
	//自动加载配置文件
	go func() {
		for {
			now := time.Now()
			// 计算下一个零点
			next := now.Add(time.Hour * 24)
			next = time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, next.Location())
			t := time.NewTimer(next.Sub(now.Add(time.Minute)))
			<-t.C

			fileRolling()
			time.Sleep(time.Minute * 1)
		}
	}()
	ticker := time.NewTicker(time.Second * 5)
	go func() {
		for {
			listenFileSize()
			<-ticker.C
		}
	}()
	running = true
}

func fileRolling() {
	if cfg == nil {
		return
	}
	for key, c := range cfg.LoggerCfgs {
		name := cfg.Root + c.File
		backupLog(name, c)
		GetLogger(key).refreshWriter(cfg.Root, c)
	}
}

func listenFileSize() {
	if cfg == nil {
		return
	}
	for key, c := range cfg.LoggerCfgs {
		name := cfg.Root + c.File
		fileInfo, err := os.Stat(name)
		if err != nil {
			continue
		}
		size := fileInfo.Size()
		if size >= getRealSize(c.MaxFileSize) {
			backupLog(name, c)
			GetLogger(key).refreshWriter(cfg.Root, c)
		}
	}
}

func backupLog(name string, c *LoggerCfg) {
	nextInfo := getNextInfo(c.File, c.MaxBackupIndex, c.Compress)
	bkName := cfg.Root + c.File + "." + nextInfo.Datestr + "." + strconv.Itoa(nextInfo.Index)
	err := os.Rename(name, bkName)
	os.Create(name)
	if err != nil {
		log.Print(err.Error())
	}
	if c.Compress {
		zipFile(bkName)
	}
	for _, file := range nextInfo.Deleting {
		os.Remove(cfg.Root + file)
	}
}

func zipFile(name string) {
	err := CompressFile(name+".zip", name)
	if err != nil {
		log.Println(err.Error())
		return
	}
	os.Remove(name)
}

//压缩文件Src到Dst
func CompressFile(Dst string, Src string) error {
	newfile, err := os.Create(Dst)
	if err != nil {
		return err
	}
	defer newfile.Close()

	file, err := os.Open(Src)
	if err != nil {
		return err
	}

	zw := gzip.NewWriter(newfile)

	filestat, err := file.Stat()
	if err != nil {
		return nil
	}

	zw.Name = filestat.Name()
	zw.ModTime = filestat.ModTime()
	_, err = io.Copy(zw, file)
	if err != nil {
		return nil
	}

	zw.Flush()
	if err := zw.Close(); err != nil {
		return nil
	}
	return nil
}

type NextInfo struct {
	Deleting []string
	Datestr  string
	Index    int
}

type LogFileInfo struct {
	Datestr string
	Index   int
	Name    string
}

func getNextInfo(fileName string, maxIndex int, compress bool) *NextInfo {
	now := time.Now()
	datestr := now.Format("2006-01-02")
	logsList := []*LogFileInfo{}
	files, _ := ioutil.ReadDir(cfg.Root)
	for _, file := range files {
		name := file.Name()
		if strings.HasPrefix(name, fileName) {
			val := strings.Split(name, ".")
			if len(val) < 4 {
				continue
			}
			index, _ := strconv.Atoi(val[3])
			info := &LogFileInfo{
				Datestr: val[2],
				Index:   index,
				Name:    name,
			}
			logsList = append(logsList, info)
		}
	}
	sort.Slice(logsList, func(i, j int) bool {
		ei := logsList[i]
		ej := logsList[j]
		if ei.Datestr == ej.Datestr {
			return ei.Index > ej.Index
		}
		return ei.Datestr > ej.Datestr
	})
	length := len(logsList)

	next := &NextInfo{
		Deleting: []string{},
		Datestr:  datestr,
		Index:    0,
	}

	if length == 0 {
		return next
	}
	newest := logsList[0]
	if newest.Datestr == datestr {
		next.Index = newest.Index + 1
	}
	for i := maxIndex - 1; i < length; i++ {
		next.Deleting = append(next.Deleting, logsList[i].Name)
	}
	return next
}

func isFileExists(name string) bool {
	_, err := os.Stat(name)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func getRealSize(s string) int64 {
	s = strings.ToUpper(s)
	str := ""
	multiple := 1
	if strings.HasSuffix(s, "KB") {
		str = "KB"
		multiple = 1024
	} else if strings.HasSuffix(s, "MB") {
		str = "MB"
		multiple = 1024 * 1024
	} else if strings.HasSuffix(s, "GB") {
		str = "KB"
		multiple = 1024 * 1024 * 1024
	}

	s = strings.Replace(s, str, "", 1)
	val, err := strconv.Atoi(s)
	if nil != err {
		return 0
	}
	return int64(val * multiple)
}
