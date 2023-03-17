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
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/denisbrodbeck/machineid"
	"github.com/dgrijalva/jwt-go"
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
	resp_code := http.StatusRequestTimeout
	msg := ""
	resp_body := ""

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

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	if resp != nil {
		resp_code = resp.StatusCode
	}
	if err != nil || resp == nil {
		msg = fmt.Sprintf("failed to send HTTP request: %v", err.Error())
		log.Error(msg)
		return resp_code, resp_body, err
	}
	defer resp.Body.Close()

	resp_code = resp.StatusCode
	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			msg = fmt.Sprintf("failed to read response body: %v", err.Error())
			log.Error(msg)
		} else {
			resp_body = string(bodyBytes)
		}
	}

	// 返回响应状态码、响应体和错误信息
	return resp_code, resp_body, err
}

func FetchURLWithRetry(verb string, url string, payload string, headers map[string]string, retries int, retryInterval time.Duration) (int, string, error) {
	resp_code := http.StatusRequestTimeout
	msg := ""
	resp_body := ""

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

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: retryInterval * time.Second,
	}

	var resp *http.Response
	for i := 0; i < retries; i++ {
		resp, err = client.Do(req)
		if resp != nil {
			resp_code = resp.StatusCode
		}
		if err == nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()

			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				msg = fmt.Sprintf("failed to read response body: %v", err.Error())
				log.Error(msg)
			} else {
				resp_body = string(bodyBytes)
				// 返回响应状态码、响应体和错误信息
				return resp_code, resp_body, nil
			}
		}
		time.Sleep(retryInterval)
	}
	if err != nil {
		return resp_code, "", err
	}
	return resp_code, "", fmt.Errorf("failed to load URL %s after %d retries: status code %d", url, retries, resp.StatusCode)
}

func GenerateAccessToken(app_key string, email string) (string, error) {
	var err error
	var device_info string
	var current_user *user.User

	// get user id
	current_user, err = user.Current()
	if err != nil {
		log.Error("Failed to get current user information: %v", err.Error())
		return "", err
	}

	// get device info
	if len(app_key) > 0 {
		device_info, err = machineid.ProtectedID(app_key)
	} else {
		device_info, err = machineid.ID()
	}
	if err != nil {
		log.Error("Failed to get device information: %v", err.Error())
		return "", err
	}

	claims := jwt.MapClaims{}
	claims["authorized"] = true

	claims["uid"] = current_user.Uid
	claims["username"] = current_user.Username
	claims["email"] = email
	claims["device_info"] = device_info

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(app_key))
}

func ExtractAccessInfo(app_key string, access_token string) (jwt.MapClaims, error) {
	if access_token == "" {
		log.Error("Blank access token in ExtractAccessInfo")
		return nil, nil
	}

	token, err := jwt.Parse(access_token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(app_key), nil
	})
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("Invalid access token")
}
