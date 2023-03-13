package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"robinmin.net/tools/xally/cmd/client/chatbot"
	"robinmin.net/tools/xally/config"
	"robinmin.net/tools/xally/shared/utility"
)

var (
	help              bool
	config_file       string
	chat_history_path string
	language          string
	command           string
	role              string
	verbose           bool
)

func init() {
	// setup flags
	flag.BoolVar(&help, "h", false, "show the help message")
	flag.StringVar(&config_file, "f", "", "config file")
	flag.StringVar(&chat_history_path, "d", "", "specify chat history path")
	flag.StringVar(&language, "p", "", "language preference, so far only support CN, JP and EN")
	flag.StringVar(&command, "c", "", "command for single line instruction")
	flag.StringVar(&role, "r", "", "default role for command")
	flag.BoolVar(&verbose, "v", false, "show detail information")

	// change the default useage
	flag.Usage = usage
}

func usage() {
	fmt.Fprintf(os.Stderr, `xally version: xally/%s
Usage: xally [-hl] [-f config_file] [-c command] [-r role] [-d history_path] [-p language_preference]

Options:
`, config.Version)
	flag.PrintDefaults()
}

func main() {
	// parse command line arguments and show help only if specified
	flag.Parse()
	if help {
		flag.Usage()
		return
	}

	// load configuration
	var err error
	if _, err = config.LoadConfig(config_file, verbose); err != nil {
		fmt.Println(err)
	}

	// update configuration by specified arguments
	if len(role) > 0 {
		config.MyConfig.System.DefaultRole = role
	}

	if len(chat_history_path) > 0 {
		config.MyConfig.System.ChatHistoryPath = chat_history_path
	}

	if len(language) > 0 {
		config.MyConfig.System.PeferenceLanguage = strings.ToUpper(language)
	}

	// output before the log mechanism works
	if verbose {
		fmt.Println("Log folder: ", config.MyConfig.System.LogPath)
		fmt.Println("Chat history folder: ", config.MyConfig.System.ChatHistoryPath)
	}

	// initialize log files
	logger := utility.NewLog(config.MyConfig.System.LogPath, config.AppName, config.MyConfig.System.LogLevel)
	defer logger.Close()
	log.Debug("System initializing......")

	bot := chatbot.NewChatbot(
		config.MyConfig.System.ChatHistoryPath,
		config.AppName,
		role,
		len(config.MyConfig.System.ChatHistoryPath) > 0,
	)
	defer bot.Close()

	if len(command) == 0 {
		bot.Run()
	} else {
		commandFields := strings.Fields(command)
		result, err := bot.CommandProcessor(command, commandFields)
		if err != nil {
			log.Error(err.Error())
		}
		log.Debug(result)
	}

	log.Debug("Quit System......")
}
