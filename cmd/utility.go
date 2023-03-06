package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/glamour"
	log "github.com/sirupsen/logrus"
	"xhqb.com/tools/xally/config"
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

func GetCurrPath() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	index := strings.LastIndex(path, string(os.PathSeparator))
	ret := path[:index]
	return ret
}

// 翻译函数，接受要翻译的文本和目标语言代码，返回翻译结果和错误信息
func translate(text string, lang string) (string, error) {
	msg := ""

	// 设置DeepL API的URL和API密钥
	apiKey := os.Getenv("DEEPL_API_KEY")
	if apiKey == "" {
		msg = config.Text("error_no_deepl_key")
		return msg, nil
	}
	apiUrl := "https://api-free.deepl.com/v2/translate"

	// 构建HTTP请求
	values := url.Values{}
	values.Set("auth_key", apiKey)
	values.Set("text", text)
	values.Set("target_lang", lang)
	req, err := http.NewRequest("POST", apiUrl,
		ioutil.NopCloser(strings.NewReader(values.Encode())))
	if err != nil {
		msg = "Failed to create HTTP request object: " + err.Error()
		return msg, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 发送HTTP请求并解析响应
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		msg = "Failed to do HTTP request: " + err.Error()
		return msg, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		msg = "Failed to read HTTP response body: " + err.Error()
		return msg, err
	}

	// 解析响应并提取翻译结果
	var result struct {
		Translations []struct {
			Text string `json:"text"`
		} `json:"translations"`
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		msg = "Failed to unmarshal HTTP response body: " + err.Error()
		return msg, err
	}

	return result.Translations[0].Text, nil
}

// 查询字典函数，接受要查询的单词和目标语言代码，返回查询结果和错误信息
func lookup(text string, lang string) (string, error) {
	msg := ""

	// 设置DeepL API的URL和API密钥
	apiKey := os.Getenv("DEEPL_API_KEY")
	if apiKey == "" {
		msg = config.Text("error_no_deepl_key")
		return msg, nil
	}
	apiUrl := "https://api-free.deepl.com/v2/lexicon"

	// 构建HTTP请求
	values := url.Values{}
	values.Set("auth_key", apiKey)
	values.Set("text", text)
	values.Set("target_lang", lang)
	req, err := http.NewRequest("POST", apiUrl,
		ioutil.NopCloser(strings.NewReader(values.Encode())))
	if err != nil {
		msg = "Failed to create HTTP request object: " + err.Error()
		return msg, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 发送HTTP请求并解析响应
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		msg = "Failed to do HTTP request: " + err.Error()
		return msg, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		msg = "Failed to read HTTP response body: " + err.Error()
		return msg, err
	}

	// 解析响应并提取查询结果
	var result struct {
		Lexemes []struct {
			Lemma string `json:"lemma"`
			Pos   string `json:"pos"`
			Sense string `json:"sense"`
		} `json:"lexemes"`
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		msg = "Failed to unmarshal HTTP response body: " + err.Error()
		return msg, err
	}

	// 构建查询结果字符串
	var sb strings.Builder
	for _, lexeme := range result.Lexemes {
		sb.WriteString(fmt.Sprintf("%s (%s): %s\n", lexeme.Lemma,
			lexeme.Pos, lexeme.Sense))
	}

	return sb.String(), nil
}
