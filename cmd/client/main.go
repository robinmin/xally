package main

import (
	"flag"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"xhqb.com/tools/xally/cmd"
	"xhqb.com/tools/xally/config"
)

var (
	help              bool
	flag_log_history  bool
	chat_history_path string
)

func init() {
	// setup flags
	flag.BoolVar(&help, "h", false, "show the help message")
	flag.BoolVar(&flag_log_history, "l", config.LogConversationHistory, "flag to log history")
	flag.StringVar(&chat_history_path, "p", config.ChatHistoryPath, "specify chat history path")

	// change the default useage
	flag.Usage = usage
}

func usage() {
	fmt.Fprintf(os.Stderr, `xally version: xally/%s
Usage: xally [-hl] [-p history_path]

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

	// initialize log files
	lg := cmd.NewLog("logs", config.AppName, log.DebugLevel)
	defer lg.Close()
	log.Debug("System initializing......")

	pwd, _ := os.Getwd()
	log.Debug("GetCurrPath = ", cmd.GetCurrPath())
	log.Debug("PWD = ", pwd)

	bot := cmd.NewChatbot(config.ChatHistoryPath, config.AppName, config.LogConversationHistory)
	defer bot.Close()

	bot.Run()
	log.Debug("Quit System......")
}
