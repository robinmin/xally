package config

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"syscall"
)

// /////////////////////////////////////////////////////////////////////////////
func IsWritable(path string, verbose bool) (isWritable bool, err error) {
	isWritable = false
	info, err := os.Stat(path)
	if err != nil {
		if verbose {
			fmt.Println("Path doesn't exist")
		}
		return
	}

	err = nil
	if !info.IsDir() {
		if verbose {
			fmt.Println("Path isn't a directory")
		}
		return
	}

	// Check if the user bit is enabled in file permission
	if info.Mode().Perm()&(1<<(uint(7))) == 0 {
		if verbose {
			fmt.Println("Write permission bit is not set on this file for user")
		}
		return
	}

	var stat syscall.Stat_t
	if err = syscall.Stat(path, &stat); err != nil {
		if verbose {
			fmt.Println("Unable to get stat")
		}
		return
	}

	err = nil
	if uint32(os.Geteuid()) != stat.Uid {
		isWritable = false
		if verbose {
			fmt.Println("User doesn't have permission to write to this directory")
		}
		return
	}

	isWritable = true
	return
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
	}

	return MyConfig, nil
}

func Text(str_key string) string {
	msg := ""
	switch MyConfig.System.PeferenceLanguage {
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
