package main

import (
	"os"

	log "github.com/sirupsen/logrus"

	"robinmin.net/tools/xally/cmd"
	"robinmin.net/tools/xally/config"
)

/******************************************************************************
*
* Entry point for x-ally client
*
*******************************************************************************/
func main() {
	// initialize log files
	lg := cmd.NewLog("logs", config.AppName, "debug")
	defer lg.Close()
	log.Debug("System initializing......")

	pwd, _ := os.Getwd()
	log.Debug("GetCurrPath = ", cmd.GetCurrPath())
	log.Debug("PWD = ", pwd)
}
