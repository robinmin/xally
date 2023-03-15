package config

import (
	"errors"
	"fmt"
	"os"
	"path"

	"gopkg.in/yaml.v3"
)

// Real constants and can not be changed
const ServerName = "X-Ally-Server"
const ServerVersion = "0.0.5"

type ProxyRoute struct {
	Name    string `yaml:"name"`
	Context string `yaml:"context"`
	Target  string `yaml:"target"`
}

type ServerConfigItems struct {
	DSN                      string       `yaml:"DSN"`
	OpenaiApiKey             string       `yaml:"openai_api_key"`
	OpenaiOrgID              string       `yaml:"openai_org_id"`
	AppToken                 string       `yaml:"app_token"`
	AppTokenLifespan         uint32       `yaml:"app_token_lifespan"`
	ListenAddr               string       `yaml:"listen_addr"`
	WhiteListRefreshInterval int64        `yaml:"white_list_refresh_interval,omitempty"`
	Routes                   []ProxyRoute `yaml:"routes,omitempty"`
}

type ServerConfig struct {
	Server *ServerConfigItems `yaml:"server"`
}

func (cfg *ServerConfig) DumpIntoYAML(cfg_file string) (string, error) {
	yaml_data, err := yaml.Marshal(cfg)
	if err != nil {
		return "", err
	}

	if err = os.WriteFile(cfg_file, yaml_data, 0644); err != nil {
		return "", err
	}
	return string(yaml_data), nil
}

func (cfg *ServerConfig) LoadFromYAML(cfg_file string) error {
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

var SvrConfig *ServerConfig

func LoadServerConfig(config_file string, verbose bool) (*ServerConfig, error) {
	SvrConfig = &ServerConfig{}

	var temp_file string
	var default_config_file string
	var err error
	if config_file == "" {
		var dir_home string
		var err error
		if dir_home, err = FindHomeDir(verbose); err != nil {
			return nil, errors.New("Failed to find home directory")
		}
		default_config_file = path.Join(dir_home, "xally_server.yaml")
		temp_file = default_config_file
	} else {
		temp_file = config_file
	}

	// Create config structure
	skip_reload := false
	if _, err = os.Stat(temp_file); os.IsNotExist(err) {
		// generate default config file
		if _, err = SvrConfig.DumpIntoYAML(temp_file); err != nil {
			if verbose {
				fmt.Println("Failed to write YAML data into :", temp_file)
			}
			temp_file = default_config_file
		} else {
			skip_reload = true
		}
	}

	// // Open config file
	if !skip_reload {
		if verbose {
			fmt.Println("Loading config file from ", temp_file)
		}

		if err = SvrConfig.LoadFromYAML(temp_file); err != nil {
			if verbose {
				fmt.Println("Failed to load configuration from config file : ", temp_file)
				fmt.Println(err)
			}
			return SvrConfig, err
		}
		// update key from env var in case of blank
		if SvrConfig.Server.OpenaiApiKey == "" {
			SvrConfig.Server.OpenaiApiKey = os.Getenv("OPENAI_API_KEY")
		}
	}

	return SvrConfig, nil
}
