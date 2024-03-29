package config

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
)

// /////////////////////////////////////////////////////////////////////////////
func IsWritable(path string, verbose bool) (isWritable bool, err error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if verbose {
			fmt.Printf("Path %s does not exist\n", path)
		}
		return false, err
	}

	file, err := os.OpenFile(path+"/.temp", os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		if verbose {
			fmt.Printf("Path %s is not writable\n", path)
		}
		return false, err
	}
	defer func() {
		file.Close()
		os.Remove(path + "/.temp")
	}()

	if verbose {
		fmt.Printf("Path %s is writable\n", path)
	}
	return true, nil
}

func GetRealFullPath(path string) (string, error) {
	if path == "" {
		return "", errors.New("Invalid filepath")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", err
	}

	temp_path, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	targetPath, err := os.Readlink(temp_path)
	if err != nil {
		return "", err
	}

	absTargetPath, err := filepath.Abs(filepath.Join(filepath.Dir(temp_path), targetPath))
	if err != nil {
		return "", err
	}

	return absTargetPath, nil
}

func FindHomeDir(verbose bool) (string, error) {
	var err error

	default_config_file := os.Getenv("HOME")
	if default_config_file == "" {
		default_config_file = "~"
	}
	default_config_file = path.Join(default_config_file, ".xally")

	if _, err = os.Stat(path.Dir(default_config_file)); os.IsNotExist(err) {
		if err = os.MkdirAll(path.Dir(default_config_file), 0755); err != nil {
			if verbose {
				fmt.Println("Failed to create main folder under home directory.")
			}
			return "", err
		}
	}

	return default_config_file, nil
}

func LoadClientConfig(config_file string, verbose bool) (*SysConfig, error) {
	var temp_file string
	var default_config_file string
	var err error

	if config_file == "" {
		var dir_home string
		var err error
		if dir_home, err = FindHomeDir(verbose); err != nil {
			return nil, errors.New("Failed to find home directory")
		}
		default_config_file = path.Join(dir_home, "xally.yaml")
		temp_file = default_config_file
	} else {
		temp_file = config_file
	}

	// Create config structure
	MyConfig = NewSysConfig(temp_file)
	skip_reload := false
	if _, err = os.Stat(temp_file); os.IsNotExist(err) {
		// generate default config file
		if _, err = MyConfig.DumpIntoYAML(temp_file); err != nil {
			if verbose {
				fmt.Println("Failed to write YAML data into :", temp_file)
			}
			temp_file = default_config_file
		} else {
			skip_reload = true
		}
	}

	// Open config file
	if !skip_reload {
		if verbose {
			fmt.Println("Loading config file from ", temp_file)
		}

		if err = MyConfig.LoadFromYAML(temp_file); err != nil {
			if verbose {
				fmt.Println("Failed to load configuration from config file : ", temp_file)
				fmt.Println(err)
			}
			return MyConfig, err
		}
		// update key from env var in case of blank
		if MyConfig.System.OpenaiApiKey == "" {
			MyConfig.System.OpenaiApiKey = os.Getenv("OPENAI_API_KEY")
		}
	}

	SetupPeferenceLanguage(MyConfig.System.PeferenceLanguage)

	return MyConfig, nil
}

var peference_language string

func SetupPeferenceLanguage(language string) {
	peference_language = language
}

func Text(str_key string) string {
	var msg string

	switch peference_language {
	case "EN", "en_US.UTF-8", "C":
		if str_val, ok := i18n_str_table_en[str_key]; ok {
			msg = str_val
		} else {
			msg = str_key
		}
	case "JP", "ja_JP.UTF-8":
		if str_val, ok := i18n_str_table_jp[str_key]; ok {
			msg = str_val
		} else {
			msg = str_key
		}
	case "CN", "zh_CN.UTF-8":
		fallthrough
	default:
		if str_val, ok := i18n_str_table[str_key]; ok {
			msg = str_val
		} else {
			msg = str_key
		}
	}
	return msg
}

func GetAcceptLanguage() string {
	switch peference_language {
	case "EN", "en_US.UTF-8", "C":
		return "en"
	case "JP", "ja_JP.UTF-8":
		return "ja"
	case "CN", "zh_CN.UTF-8":
		return "zh"
	default:
		return "en"
	}
}
