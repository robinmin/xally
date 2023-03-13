package main

import (
	"os"

	log "github.com/sirupsen/logrus"

	"robinmin.net/tools/xally/config"
	"robinmin.net/tools/xally/shared/utility"
)

/******************************************************************************
*
* Entry point for x-ally client
*
*******************************************************************************/
func main() {
	// initialize log files
	lg := utility.NewLog("logs", config.AppName, "debug")
	defer lg.Close()
	log.Debug("System initializing......")

	pwd, _ := os.Getwd()
	log.Debug("GetCurrPath = ", utility.GetCurrPath())
	log.Debug("PWD = ", pwd)
}
