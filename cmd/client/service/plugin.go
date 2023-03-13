package service

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	log "github.com/sirupsen/logrus"

	"github.com/robinmin/xally/shared/utility"
)

// const RX_FILE_NAME = "^[^\\\\/?%*:|\"<>]+(\\.[^\\\\/?%*:|\"<>]+)*$"
// const RX_WEB_EMAIL = `^\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*$`
// const RX_WEB_URL = `^http(s)://([\w-]+\.)+[\w-]+(/[\w-./?%&=]*)?$`
// const RX_WEB_DOMAIN = `[a-zA-Z0-9][-a-zA-Z0-9]{0,62}(\.[a-zA-Z0-9][-a-zA-Z0-9]{0,62})+\.?`
// const RX_WEB_MOBILE = `^(13[0-9]|14[01456879]|15[0-35-9]|16[2567]|17[0-8]|18[0-9]|19[0-35-9])\d{8}$`
// const RX_WEB_CNID = `(^\d{15}$)|(^\d{18}$)|(^\d{17}(\d|X|x)$)`

const PLUGIN_NAME_FILE = "file-content"
const PLUGIN_NAME_WEB_CONTENT = "web-content"
const PLUGIN_NAME_WEB_SUMMARY = "web-summary"

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
	pm.AddPlugin(&FilePlugin{})
	pm.AddPlugin(&WebSummaryPlugin{mode: PLUGIN_NAME_WEB_CONTENT})
	pm.AddPlugin(&WebSummaryPlugin{mode: PLUGIN_NAME_WEB_SUMMARY})

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
		if tmp_processed, tmp_replaced_msg, tmp_replaced_cmd, tmp_err := p.Execute(original_msg, arr_cmd); tmp_err != nil {
			log.Error(tmp_err)
		} else {
			if tmp_processed {
				processed = true
				return processed, tmp_replaced_msg, tmp_replaced_cmd, tmp_err
			}
		}
	}
	return processed, "", nil, nil
}

type FilePlugin struct {
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
		return processed, "", nil, err
	}

	// if !plugin.rx_pattern.MatchString(original_msg) {
	// 	return processed, "", nil, err
	// }

	var file_name string
	if len(arr_cmd) > 0 && arr_cmd[0] == PLUGIN_NAME_FILE {
		file_name = arr_cmd[0]
	} else {
		file_name = original_msg
	}

	if _, err = os.Stat(file_name); os.IsNotExist(err) {
		log.Debug("It's not a file : ", file_name)
		return processed, "", nil, err
	}

	var data []byte
	if data, err = os.ReadFile(file_name); err != nil {
		log.Error("Failed to read data from : ", file_name)
		return processed, "", nil, err
	}

	replaced_msg = string(data)
	replaced_cmd = append([]string{PLUGIN_NAME_FILE}, strings.Fields(replaced_msg)...)
	// stop bubble up
	processed = true

	return processed, replaced_msg, replaced_cmd, err
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
	err = nil

	if original_msg == "" {
		return processed, "", nil, err
	}

	// if !plugin.rx_pattern.MatchString(original_msg) {
	// 	return processed, "", nil, err
	// }

	var url_str string
	if len(arr_cmd) > 0 && arr_cmd[0] == PLUGIN_NAME_WEB_SUMMARY {
		url_str = arr_cmd[0]
	} else {
		url_str = original_msg
	}
	if _, err = url.Parse(url_str); err != nil {
		return processed, "", nil, err
	}

	headers := map[string]string{}
	payload := ""
	statusCode, responseBody, err := utility.FetchURL("GET", url_str, payload, headers)
	if err != nil || statusCode != 200 || responseBody == "" {
		log.Error("Failed to fetch web page from : ", url_str)
		return processed, "", nil, err
	}

	converter := md.NewConverter("", true, nil)
	replaced_msg, err = converter.ConvertString(responseBody)
	if err != nil {
		log.Error("Failed to convert content in markdown")
		return processed, "", nil, err
	}

	if plugin.mode == PLUGIN_NAME_WEB_CONTENT {
		replaced_cmd = append([]string{PLUGIN_NAME_WEB_CONTENT}, strings.Fields(replaced_msg)...)
	} else {
		replaced_cmd = append([]string{PLUGIN_NAME_WEB_SUMMARY}, strings.Fields(replaced_msg)...)
		replaced_msg = fmt.Sprintf(
			"%s\n-------------------------\n%s",
			"Can you help me extract key information based on the following and summarize them in bullet point form in my prefferd language as concisely as possible?",
			replaced_msg,
		)
	}
	// stop bubble up
	processed = true

	return processed, "", nil, err
}
