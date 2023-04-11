package utility

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"

	"github.com/charmbracelet/glamour"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"

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
	api_key := config.MyConfig.System.DeeplApiKey
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
	api_key := config.MyConfig.System.DeeplApiKey
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
	// set the http headers if not specified
	acceptType := req.Header.Get("Accept")
	if acceptType == "" {
		req.Header.Set("Accept", "application/json; charset=utf-8")
	}
	acceptLang := req.Header.Get("Accept-Language")
	if acceptLang == "" {
		req.Header.Set("Accept-Language", "application/json; charset=utf-8")
	}
	contentType := req.Header.Get("Content-Type")
	if contentType == "" {
		req.Header.Set("Content-Type", config.GetAcceptLanguage())
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/111.0.0.0 Safari/537.36")

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
	// set the http headers if not specified
	acceptType := req.Header.Get("Accept")
	if acceptType == "" {
		req.Header.Set("Accept", "application/json; charset=utf-8")
	}
	acceptLang := req.Header.Get("Accept-Language")
	if acceptLang == "" {
		req.Header.Set("Accept-Language", "application/json; charset=utf-8")
	}
	contentType := req.Header.Get("Content-Type")
	if contentType == "" {
		req.Header.Set("Content-Type", config.GetAcceptLanguage())
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/111.0.0.0 Safari/537.36")

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

func SendEmail(to, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", config.SvrConfig.Server.SMTPUsername)
	m.SetHeader("To", to)
	// m.SetAddressHeader("Cc", "dan@example.com", "Dan")
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)
	// m.Attach("/home/Alex/lolcat.jpg")

	d := gomail.NewDialer(
		config.SvrConfig.Server.SMTPServer,
		config.SvrConfig.Server.SMTPPort,
		config.SvrConfig.Server.SMTPUsername,
		config.SvrConfig.Server.SMTPPassword,
	)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	// Send the email to Bob, Cora and Dan.
	if err := d.DialAndSend(m); err != nil {
		log.Error("Failed to send email : " + err.Error())
		return err
	}

	// // 配置SMTP客户端
	// auth := smtp.PlainAuth(
	// 	"",
	// 	config.SvrConfig.Server.SMTPUsername,
	// 	config.SvrConfig.Server.SMTPPassword,
	// 	config.SvrConfig.Server.SMTPServer,
	// )
	// addr := fmt.Sprintf("%s:%d", config.SvrConfig.Server.SMTPServer, config.SvrConfig.Server.SMTPPort)

	// // 构造邮件内容
	// msg := []byte("To: " + to + "\r\n" +
	// 	"Subject: " + subject + "\r\n" +
	// 	"\r\n" +
	// 	body + "\r\n")

	// // 发送邮件
	// err := smtp.SendMail(addr, auth, config.SvrConfig.Server.SMTPUsername, []string{to}, msg)
	// if err != nil {
	// 	return err
	// }

	return nil
}

func GetBaseURL(endpoint_url string) string {
	u, err := url.Parse(endpoint_url)
	if err == nil {
		var parent *url.URL
		if len(u.Path) > 1 {
			parent = &url.URL{
				Scheme: u.Scheme,
				Host:   u.Host,
				Path:   path.Dir(u.Path),
			}
		} else {
			parent = &url.URL{
				Scheme: u.Scheme,
				Host:   u.Host,
				Path:   u.Path,
			}
		}
		return parent.String()
	}
	return "http://127.0.0.1:8090/"
}

func IsValidEmail(email string) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`).MatchString(email)
}

func IsValidURL(url_str string) bool {
	u, err := url.Parse(url_str)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}
	return true
}

func AcceptJSONResponse(ctx *gin.Context) bool {
	accpt_type := ctx.Request.Header.Get("Accept")
	return strings.Contains(strings.ToLower(accpt_type), "application/json")
}

// encoding determine for html page , eg: gbk gb2312 GB18030
func determineEncoding(r io.Reader) encoding.Encoding {
	bytes, err := bufio.NewReader(r).Peek(1024)
	if err != nil {
		return nil
	}
	e, _, _ := charset.DetermineEncoding(bytes, "")
	return e
}

// 将[]byte转换为UTF-8编码
func ConvertToUTF8(src []byte) ([]byte, error) {
	// convert byte slice to io.Reader
	reader := bytes.NewReader(src)
	enc := determineEncoding(reader)
	if enc == nil {
		return nil, errors.New("Failed to get encoding on provided data")
	}

	utf8Reader := transform.NewReader(reader, enc.NewDecoder())
	det, err := ioutil.ReadAll(utf8Reader)
	if err != nil {
		return nil, err
	}
	return det, nil
}
