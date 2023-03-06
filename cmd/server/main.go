package main

import (
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/robinmin/xally/cmd"
	"github.com/robinmin/xally/config"
)

/******************************************************************************
*
* Entry point for x-ally client
*
*******************************************************************************/
func main() {
	// initialize log files
	lg := cmd.NewLog("logs", config.AppName, log.DebugLevel)
	defer lg.Close()
	log.Debug("System initializing......")

	pwd, _ := os.Getwd()
	log.Debug("GetCurrPath = ", cmd.GetCurrPath())
	log.Debug("PWD = ", pwd)
}
