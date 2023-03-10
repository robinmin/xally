package utility

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

	"github.com/robinmin/xally/config"
)

type LogFile struct {
	FileHandle *os.File
	LogPath    string
	LogLevel   string
}

func GetYYYYMMDD() string {
	t := time.Now()
	return fmt.Sprintf("%04d%02d%02d", t.Year(), t.Month(), t.Day())
}

func GetYYYYMM() string {
	t := time.Now()
	return fmt.Sprintf("%04d%02d", t.Year(), t.Month())
}

func NewLog(log_path string, name string, level string) *LogFile {
	logger := &LogFile{
		LogPath:  log_path,
		LogLevel: level,
	}
	if _, err := os.Stat(logger.LogPath); os.IsNotExist(err) {
		errDir := os.MkdirAll(logger.LogPath, 0755)
		if errDir != nil {
			log.Error(err)
			return logger
		}
	}

	var err error
	log_file := logger.LogPath + "/" + name + "_" + GetYYYYMMDD() + ".log"
	if logger.FileHandle == nil {
		logger.FileHandle, err = os.OpenFile(log_file, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			log.Error(err)
			return logger
		}

		log.SetOutput(logger.FileHandle)
		// log.SetFormatter(&log.JSONFormatter{})
		var level_int log.Level
		if level_int, err = log.ParseLevel(level); err == nil {
			log.SetLevel(level_int)
		} else {
			log.SetLevel(log.DebugLevel)
		}
	}
	return logger
}

func (lf *LogFile) Close() {
	if lf.FileHandle != nil {
		lf.FileHandle.Close()
		lf.FileHandle = nil
	}
}

func EchoInfo(msg string) {
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
func Translate(text string, lang string) (string, error) {
	msg := ""

	// 设置DeepL API的URL和API密钥
	api_key := config.MyConfig.System.APIKeyDeepl
	if api_key == "" {
		api_key = os.Getenv("DEEPL_API_KEY")
		if api_key == "" {
			msg = config.Text("error_no_deepl_key")
			return msg, nil
		}
	}
	api_url := config.MyConfig.System.APIEndpointDeepl + "/translate"

	// 构建HTTP请求
	values := url.Values{}
	values.Set("auth_key", api_key)
	values.Set("text", text)
	values.Set("target_lang", lang)
	req, err := http.NewRequest("POST", api_url,
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
func Lookup(text string, lang string) (string, error) {
	msg := ""

	// 设置DeepL API的URL和API密钥
	api_key := config.MyConfig.System.APIKeyDeepl
	if api_key == "" {
		api_key = os.Getenv("DEEPL_API_KEY")
		if api_key == "" {
			msg = config.Text("error_no_deepl_key")
			return msg, nil
		}
	}
	api_url := config.MyConfig.System.APIEndpointDeepl + "/lexicon"

	// 构建HTTP请求
	values := url.Values{}
	values.Set("auth_key", api_key)
	values.Set("text", text)
	values.Set("target_lang", lang)
	req, err := http.NewRequest("POST", api_url,
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

func FetchURL(verb string, url string, payload string, headers map[string]string) (int, string, error) {
	resp_code := http.StatusOK
	msg := ""
	resp_body := ""

	// 创建HTTP客户端
	client := &http.Client{}

	// 创建HTTP请求
	req, err := http.NewRequest(verb, url, bytes.NewBuffer([]byte(payload)))
	if err != nil {
		msg = fmt.Sprintf("failed to create HTTP request: %v", err.Error())
		log.Error(msg)
		return resp_code, resp_body, err
	}

	// 设置HTTP请求头
	if headers != nil && len(headers) > 0 {
		for key, value := range headers {
			req.Header.Set(key, value)
		}
	}

	// 发送HTTP请求
	resp, err := client.Do(req)
	if err != nil {
		msg = fmt.Sprintf("failed to send HTTP request: %v", err.Error())
		log.Error(msg)
		return resp_code, resp_body, err
	}
	defer resp.Body.Close()

	resp_code = resp.StatusCode
	if resp.StatusCode == http.StatusOK {
		// 读取响应体
		bodyBytes, err1 := io.ReadAll(resp.Body)
		if err1 == nil {
			resp_body = string(bodyBytes)
		} else {
			msg = fmt.Sprintf("failed to read response body: %v", err.Error())
			log.Error(msg)
		}
	} else {
		msg = fmt.Sprintf("Invalid response code : : %v", resp.StatusCode)
		log.Error(msg)
		err = errors.New(msg)
	}

	// 返回响应状态码、响应体和错误信息
	return resp_code, resp_body, nil
}
