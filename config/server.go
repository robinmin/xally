package config

// Real constants and can not be changed
const ServerName = "X-Ally-Server"
const ServerVersion = "0.0.5"

type ServerConfig struct {
	DSN        string
	ServerPort string

	TokenApiSecret string
	TokenLifespan  uint32
}

var SvrConfig *ServerConfig

func LoadServerConfig(config_file string, verbose bool) (*ServerConfig, error) {
	SvrConfig = &ServerConfig{
		DSN:        "robin:dragon2001@tcp(127.0.0.1:3306)/xally?charset=utf8&parseTime=True&loc=Local",
		ServerPort: ":8080",
	}

	return SvrConfig, nil
}
