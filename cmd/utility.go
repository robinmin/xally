package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/glamour"
	log "github.com/sirupsen/logrus"
)

type LogFile struct {
	file_handle *os.File
	log_path    string
	level       log.Level
}

func get_yyyymmdd() string {
	t := time.Now()
	return fmt.Sprintf("%04d%02d%02d", t.Year(), t.Month(), t.Day())
}

func NewLog(log_path string, name string, level log.Level) *LogFile {
	lf := &LogFile{
		log_path: log_path,
		level:    level,
	}
	if _, err := os.Stat(lf.log_path); os.IsNotExist(err) {
		errDir := os.MkdirAll(lf.log_path, 0755)
		if errDir != nil {
			log.Error(err)
			return lf
		}
	}

	var err error
	log_file := lf.log_path + "/" + name + "_" + get_yyyymmdd() + ".log"
	if lf.file_handle == nil {
		lf.file_handle, err = os.OpenFile(log_file, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			log.Error(err)
			return lf
		}

		log.SetOutput(lf.file_handle)
		// log.SetFormatter(&log.JSONFormatter{})
		log.SetLevel(lf.level)

	}
	return lf
}

func (lf *LogFile) Close() {
	if lf.file_handle != nil {
		lf.file_handle.Close()
		lf.file_handle = nil
	}
}

func echo_info(msg string) {
	if len(msg) > 0 {
		out, _ := glamour.Render(msg, "dark")
		fmt.Print(out)
	}
}

// func echo_info_with_style(msg string) {
// 	if len(msg) > 0 {
// 		rendered, _ := glamour.RenderWithEnvironmentConfig(msg)
// 		fmt.Println(rendered)
// 	}
// }

func GetCurrPath() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	index := strings.LastIndex(path, string(os.PathSeparator))
	ret := path[:index]
	return ret
}
