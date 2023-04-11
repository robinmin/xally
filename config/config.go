package config

import (
	"fmt"
	"os"
	"path"
	"strings"

	"gopkg.in/yaml.v3"
)

// Real constants and can not be changed
const AppName = "X-Ally"
const Version = "0.1.9"

// const MaxTokens = 4096
const PROXY_TOKEN_NAME = "X-ALLY-TOKEN"

type SysRole struct {
	Name        string  `yaml:"name,omitempty"`
	Model       string  `yaml:"model,omitempty"`
	Avatar      string  `yaml:"avatar,omitempty"`
	Temperature float32 `yaml:"temperature,omitempty"`
	TopP        int     `yaml:"top_p,omitempty"`
	Prompt      string  `yaml:"prompt,omitempty"`
	Opening     string  `yaml:"opening,omitempty"`
}

type SysSystem struct {
	SentryDSN         string `yaml:"sentry_dsn,omitempty"`
	ChatHistoryPath   string `yaml:"chat_history_path,omitempty"`
	LogPath           string `yaml:"log_path,omitempty"`
	LogLevel          string `yaml:"log_level,omitempty"`
	PeferenceLanguage string `yaml:"peference_language,omitempty"`
	DefaultRole       string `yaml:"default_role,omitempty"`
	APIEndpointOpenai string `yaml:"api_endpoint_openai,omitempty"`
	APIEndpointDeepl  string `yaml:"api_endpoint_deepl,omitempty"`
	// APIKeyOpenai      string `yaml:"api_key_openai,omitempty"`
	OpenaiApiKey string `yaml:"openai_api_key"`
	OpenaiOrgID  string `yaml:"openai_org_id"`
	DeeplApiKey  string `yaml:"deepl_api_key,omitempty"`

	// APIOrgIDOpenai string `yaml:"api_orgid_openai,omitempty"`
	UseSharedMode uint32 `yaml:"use_shared_mode,omitempty"`
	AppToken      string `yaml:"app_token,omitempty"`
	Email         string `yaml:"email,omitempty"`

	DebugMode bool `yaml:"debug_mode,omitempty"`
}

type SysConfig struct {
	System SysSystem `yaml:"system"`

	Roles map[string]SysRole `yaml:"roles"`
}

var MyConfig *SysConfig

func NewSysConfig(cfg_file string) *SysConfig {
	cfg := &SysConfig{
		System: SysSystem{
			SentryDSN:         "",
			ChatHistoryPath:   path.Dir(cfg_file),
			LogPath:           path.Dir(cfg_file),
			LogLevel:          "info",
			PeferenceLanguage: os.Getenv("LANG"),
			DefaultRole:       "fullstack",
			APIEndpointOpenai: "https://api.openai.com/v1",
			APIEndpointDeepl:  "https://api-free.deepl.com/v2",
			OpenaiApiKey:      "",
			OpenaiOrgID:       "",
			DeeplApiKey:       "",

			// APIOrgIDOpenai: "",
			UseSharedMode: 0,
			AppToken:      "",
			Email:         "",
			DebugMode:     false,
		},
		Roles: map[string]SysRole{
			"expert": {
				Name:        "expert",
				Avatar:      "🐬",
				Temperature: 0.2,
				TopP:        1,
				Prompt:      `You are ChatGPT, a large language model trained by OpenAI. Answer as concisely as possible.`,
				Opening:     "",
			},
			"assistant": {
				Name:        "assistant",
				Avatar:      "🧰",
				Temperature: 0.2,
				TopP:        1,
				Prompt:      `You are a ChatGPT-based daily chit-chat bot with answers that are as concise and soft as possible..`,
				Opening:     "",
			},
			"architect": {
				Name:        "architect",
				Model:       "gpt-3.5-turbo",
				Avatar:      "🏡",
				Temperature: 0.2,
				TopP:        1,
				Prompt:      `I want you to act as an IT Architect. I will provide some details about the functionality of an application or other digital product, and it will be your job to come up with ways to integrate it into the IT landscape. This could involve analyzing business requirements, performing a gap analysis and mapping the functionality of the new system to the existing IT landscape. Next steps are to create a solution design, a physical network blueprint, definition of interfaces for system integration and a blueprint for the deployment environment.`,
			},
			"tester": {
				Name:        "tester",
				Avatar:      "⁉",
				Temperature: 0.2,
				TopP:        1,
				Prompt:      `I want you to act as a software quality assurance tester for a new software application. Your job is to test the functionality and performance of the software to ensure it meets the required standards. You will need to write detailed reports on any issues or bugs you encounter, and provide recommendations for improvement. Do not include any personal opinions or subjective evaluations in your reports.`,
			},
			"fullstack": {
				Name:        "fullstack",
				Model:       "gpt-3.5-turbo",
				Avatar:      "🦄",
				Temperature: 0.2,
				TopP:        1,
				Prompt:      `I want you to act as a Senior full stack developer. I will describe the details, you will take first priority to use these tools: SQL, golang, Spring Boot. Code it in production level as concisely as possible."`,
			},
			"frontend": {
				Name:        "frontend",
				Model:       "gpt-3.5-turbo",
				Avatar:      "🍏",
				Temperature: 0.2,
				TopP:        1,
				Prompt:      "I want you to act as a Senior Frontend developer. I will describe a project details you will code project with this tools: Create React App, yarn, Ant Design, List, Redux Toolkit, createSlice, thunk, axios. You should merge files in single index.js file and nothing else. Do not write explanations.",
			},
			"backend": {
				Name:        "backend",
				Model:       "gpt-3.5-turbo",
				Avatar:      "🌲",
				Temperature: 0.2,
				TopP:        1,
				Prompt:      "I want you to act as a Senior Backend developer. I will describe a project details you will code project with this tools: SQL, golang, Spring Boot. Do not write explanations.",
			},
			"frontend_mini": {
				Name:        "frontend_mini",
				Avatar:      "🐸",
				Temperature: 0.2,
				TopP:        1,
				Prompt:      "I want you act as a Senior wechat mini programer. I will descript the request, then you need to generate code in Taro with necesary comments. Please proccess all kinds of fails and exceptions in production level.",
			},
			"translator_en": {
				Name:        "translator_en",
				Avatar:      "🍭",
				Temperature: 0.2,
				TopP:        1,
				Prompt:      `I want you to act as an English translator, spelling corrector and improver. I will speak to you in any language and you will detect the language, translate it and answer in the corrected and improved version of my text, in English. I want you to replace my simplified A0-level words and sentences with more beautiful and elegant, upper level English words and sentences. Keep the meaning same, but make them more literary. I want you to only reply the correction, the improvements and nothing else, do not write explanations.`,
			},
		},
	}
	if cfg.System.PeferenceLanguage == "" {
		cfg.System.PeferenceLanguage = "CN"
	}
	return cfg
}

func (cfg *SysConfig) IsSharedMode() bool {
	return cfg.System.UseSharedMode > 0 && len(cfg.System.AppToken) > 0
}

func (cfg *SysConfig) GetCurrentMode(connected bool) string {
	var flags string
	if cfg.System.UseSharedMode > 0 {
		if len(cfg.System.AppToken) > 0 && connected {
			flags = "🔥"
		} else {
			flags = "❌"
		}
	} else {
		if cfg.UsingOriginalService() {
			flags = "✅"
		} else {
			flags = "🚧"
		}
	}
	if cfg.DebugMode() {
		flags = flags + "🐛"
	}
	return flags
}

func (cfg *SysConfig) DumpIntoYAML(cfg_file string) (string, error) {
	yaml_data, err := yaml.Marshal(cfg)
	if err != nil {
		return "", err
	}

	if err = os.WriteFile(cfg_file, yaml_data, 0644); err != nil {
		return "", err
	}
	return string(yaml_data), nil
}

func (cfg *SysConfig) LoadFromYAML(cfg_file string) error {
	var data []byte
	var err error

	if data, err = os.ReadFile(cfg_file); err != nil {
		return err
	}
	if err = yaml.Unmarshal(data, cfg); err != nil {
		return err
	}
	return nil
}

func (cfg *SysConfig) FindRole(role_name string) (*SysRole, error) {
	// find the role name
	role := &SysRole{}
	var ok bool
	if *role, ok = cfg.Roles[role_name]; !ok {
		// use the default role
		if *role, ok = cfg.Roles[cfg.System.DefaultRole]; !ok {
			return nil, fmt.Errorf("Invalid default role : " + cfg.System.DefaultRole)
		}
	}
	return role, nil
}

func (cfg *SysConfig) UsingOriginalService() bool {
	if "https://api.openai.com/v1" == strings.ToLower(cfg.System.APIEndpointOpenai) {
		return true
	}
	return false
}

func (cfg *SysConfig) DebugMode() bool {
	return cfg.System.DebugMode
}
