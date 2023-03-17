package service

import (
	"net/url"
	"os"
	"regexp"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	log "github.com/sirupsen/logrus"

	"github.com/robinmin/xally/config"
	"github.com/robinmin/xally/shared/utility"
)

const PLUGIN_NAME_FILE_CONTENT = "file-content"
const PLUGIN_NAME_FILE_SUMMARY = "file-summary"
const PLUGIN_NAME_FILE_TRANSLATE_CN = "file-translate-cn"
const PLUGIN_NAME_FILE_TRANSLATE_EN = "file-translate-en"
const PLUGIN_NAME_FILE_TRANSLATE_JP = "file-translate-jp"

const PLUGIN_NAME_WEB_CONTENT = "web-content"
const PLUGIN_NAME_WEB_SUMMARY = "web-summary"
const PLUGIN_NAME_WEB_TRANSLATE_CN = "web-translate-cn"
const PLUGIN_NAME_WEB_TRANSLATE_EN = "web-translate-en"
const PLUGIN_NAME_WEB_TRANSLATE_JP = "web-translate-jp"

type Plugin interface {
	// GetName() string

	Open() error
	Close() error
	Execute(original_msg string, arr_cmd []string) (processed bool, replaced_msg string, replaced_cmd []string, err error)
}

type PluginManager struct {
	// name string

	plugins []Plugin
	// plugin_map map[string]*Plugin
}

func NewPluginManager() *PluginManager {
	pm := &PluginManager{}
	pm.AddPlugin(&FilePlugin{mode: PLUGIN_NAME_FILE_CONTENT})
	pm.AddPlugin(&FilePlugin{mode: PLUGIN_NAME_FILE_SUMMARY})
	pm.AddPlugin(&FilePlugin{mode: PLUGIN_NAME_FILE_TRANSLATE_CN})
	pm.AddPlugin(&FilePlugin{mode: PLUGIN_NAME_FILE_TRANSLATE_EN})
	pm.AddPlugin(&FilePlugin{mode: PLUGIN_NAME_FILE_TRANSLATE_JP})

	pm.AddPlugin(&WebSummaryPlugin{mode: PLUGIN_NAME_WEB_CONTENT})
	pm.AddPlugin(&WebSummaryPlugin{mode: PLUGIN_NAME_WEB_SUMMARY})
	pm.AddPlugin(&WebSummaryPlugin{mode: PLUGIN_NAME_WEB_TRANSLATE_CN})
	pm.AddPlugin(&WebSummaryPlugin{mode: PLUGIN_NAME_WEB_TRANSLATE_EN})
	pm.AddPlugin(&WebSummaryPlugin{mode: PLUGIN_NAME_WEB_TRANSLATE_JP})

	return pm
}

func (pm *PluginManager) AddPlugin(p Plugin) {
	pm.plugins = append(pm.plugins, p)
	// pm.plugin_map[name] = &p
}

// func (pm *PluginManager) GetPlugin(name string) *Plugin {
// 	if plugin, ok := pm.plugin_map[name]; !ok {
// 		return plugin
// 	}
// 	return nil
// }

func (pm *PluginManager) Open() error {
	for _, p := range pm.plugins {
		err := p.Open()
		if err != nil {
			log.Error("Failed to close plugin on plugin manager")
		}
	}
	return nil
}

func (pm *PluginManager) Close() error {
	for _, p := range pm.plugins {
		err := p.Close()
		if err != nil {
			log.Error("Failed to close plugin on plugin manager")
		}
	}
	return nil
}

func (pm *PluginManager) Execute(original_msg string, arr_cmd []string) (processed bool, replaced_msg string, replaced_cmd []string, err error) {
	processed = false
	for _, p := range pm.plugins {
		tmp_processed, tmp_replaced_msg, tmp_replaced_cmd, tmp_err := p.Execute(original_msg, arr_cmd)
		if tmp_err != nil {
			log.Error(tmp_err)
		}
		if tmp_processed {
			processed = true
			return processed, tmp_replaced_msg, tmp_replaced_cmd, tmp_err
		}
	}
	return processed, "", nil, nil
}

type FilePlugin struct {
	mode string
	// rx_pattern *regexp.Regexp
}

func (plugin *FilePlugin) Open() error {
	// plugin.rx_pattern = regexp.MustCompile(RX_FILE_NAME)
	return nil
}

func (*FilePlugin) Close() error {
	return nil
}

func (plugin *FilePlugin) Execute(original_msg string, arr_cmd []string) (processed bool, replaced_msg string, replaced_cmd []string, err error) {
	processed = false
	err = nil

	if original_msg == "" {
		return
	}

	var file_name string
	var current_mode string
	if len(arr_cmd) > 0 {
		file_name = strings.Join(arr_cmd[1:], " ")
		current_mode = arr_cmd[0]
	} else {
		file_name = original_msg
		current_mode = PLUGIN_NAME_FILE_CONTENT
	}

	// ignore invalid input or skip the other modes
	if file_name == "" || plugin.mode != current_mode {
		return
	}
	// stop bubble up
	processed = true

	if _, err = os.Stat(file_name); os.IsNotExist(err) {
		log.Debug("It's not a file : ", file_name)
		return
	}

	var data []byte
	if data, err = os.ReadFile(file_name); err != nil {
		log.Error("Failed to read data from : ", file_name)
		return
	}
	if len(data) <= 0 {
		log.Info("Blank file ", file_name)
		return
	}

	replaced_msg = string(data)
	replaced_cmd = append([]string{plugin.mode}, strings.Fields(replaced_msg)...)
	var prompt_msg string
	switch current_mode {
	case PLUGIN_NAME_FILE_SUMMARY:
		prompt_msg = config.Text("prompt_content_summary")
	case PLUGIN_NAME_FILE_TRANSLATE_CN:
		prompt_msg = config.Text("prompt_translate_cn")
	case PLUGIN_NAME_FILE_TRANSLATE_EN:
		prompt_msg = config.Text("prompt_translate_en")
	case PLUGIN_NAME_FILE_TRANSLATE_JP:
		prompt_msg = config.Text("prompt_translate_jp")
	}
	if len(prompt_msg) > 0 {
		replaced_msg = prompt_msg + "\n\n-------------------------\n" + replaced_msg
	}

	return
}

type WebSummaryPlugin struct {
	mode string
}

func (plugin *WebSummaryPlugin) Open() error {
	// plugin.rx_pattern = regexp.MustCompile(RX_WEB_URL)
	return nil
}

func (*WebSummaryPlugin) Close() error {
	return nil
}

func (plugin *WebSummaryPlugin) Execute(original_msg string, arr_cmd []string) (processed bool, replaced_msg string, replaced_cmd []string, err error) {
	processed = false
	replaced_msg = ""
	err = nil

	if original_msg == "" {
		return
	}

	var url_str string
	var current_mode string
	if len(arr_cmd) > 0 {
		url_str = strings.Join(arr_cmd[1:], " ")
		current_mode = arr_cmd[0]
	} else {
		url_str = original_msg
		current_mode = PLUGIN_NAME_WEB_CONTENT
	}
	if _, err = url.Parse(url_str); err != nil || plugin.mode != current_mode {
		return
	}
	// stop bubble up
	processed = true

	headers := map[string]string{}
	payload := ""
	statusCode, responseBody, err := utility.FetchURLWithRetry("GET", url_str, payload, headers, 3, 5)
	if err != nil || statusCode != 200 || responseBody == "" {
		log.Error("Failed to fetch web page from : ", url_str)
		return
	}

	// try to remove hyper-links
	re := regexp.MustCompile(`<a href=['"](.*?)['"]>(.*?)</a>`)
	responseBody = re.ReplaceAllString(responseBody, "$2")

	converter := md.NewConverter("", true, nil)
	replaced_msg, err = converter.ConvertString(responseBody)
	if err != nil {
		log.Error("Failed to convert content in markdown")
		return
	}

	replaced_cmd = append([]string{plugin.mode}, strings.Fields(replaced_msg)...)
	var prompt_msg string
	switch plugin.mode {
	case PLUGIN_NAME_WEB_SUMMARY:
		prompt_msg = config.Text("prompt_content_summary")
	case PLUGIN_NAME_WEB_TRANSLATE_CN:
		prompt_msg = config.Text("prompt_translate_cn")
	case PLUGIN_NAME_WEB_TRANSLATE_EN:
		prompt_msg = config.Text("prompt_translate_en")
	case PLUGIN_NAME_WEB_TRANSLATE_JP:
		prompt_msg = config.Text("prompt_translate_jp")
	}
	if len(prompt_msg) > 0 {
		replaced_msg = prompt_msg + "\n-------------------------\n" + replaced_msg
	}

	return
}
