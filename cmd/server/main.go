package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/robinmin/xally/cmd/server/controller"
	"github.com/robinmin/xally/config"
	"github.com/robinmin/xally/shared/utility"
)

/******************************************************************************
*
* Entry point for x-ally server
*
*******************************************************************************/

var (
	help        bool
	config_file string
	verbose     bool
)

func init() {
	// setup flags
	flag.BoolVar(&help, "h", false, "show the help message")
	flag.StringVar(&config_file, "f", "", "config file")
	flag.BoolVar(&verbose, "v", false, "show detail information")

	// change the default useage
	flag.Usage = usage
}

func usage() {
	fmt.Fprintf(os.Stderr, `xally_server version: xally_server/%s
Usage: xally_server [-hv] [-f config_file]

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

	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// // load configuration
	var err error
	if _, err = config.LoadServerConfig(config_file, verbose); err != nil {
		fmt.Println(err)
		return
	}

	// initialize log files
	logger := utility.NewLog("logs", config.ServerName, "debug")
	defer logger.Close()

	log.Debug(config.ServerName, " initializing......")
	// 如果需要同时将日志写入文件和控制台，请使用以下代码。
	gin.DefaultWriter = io.MultiWriter(logger.FileHandle, os.Stdout)
	// 禁用控制台颜色，将日志写入文件时不需要控制台颜色。
	// gin.DisableConsoleColor()

	if len(config.SvrConfig.Server.SentryDSN) > 0 {
		utility.InitSentry(config.SvrConfig.Server.SentryDSN, false)
		defer utility.CloseSentry()

		utility.ReportEvent(utility.EVT_SERVER_INIT, "Enter Server", nil)
	}

	// 初始化gin框架
	api, router := controller.NewAPIHandler(
		config.SvrConfig.Server.AppToken,
		config.SvrConfig.Server.AppTokenLifespan,
		config.SvrConfig.Server.DBDriver,
		config.GetServerDSN(),
		verbose,
	)

	// 注册路由
	api.RegisterRoutes(router, &config.SvrConfig.Server.Routes)
	server := &http.Server{
		Addr:    config.SvrConfig.Server.ListenAddr,
		Handler: router,
	}

	// // Initializing the server in a goroutine so that
	// // it won't block the graceful shutdown handling below
	go func() {
		log.Info("Server is almost ready......")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("listen: ", err.Error())
		}
		log.Info("Server is shutting down......")
	}()

	// Listen for the interrupt signal.
	<-ctx.Done()

	// Restore default behavior on the interrupt signal and notify user of shutdown.
	stop()
	log.Info("shutting down gracefully, press Ctrl+C again to force")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown: ", err)
	}

	log.Info("Server exiting......")
	if len(config.SvrConfig.Server.SentryDSN) > 0 {
		utility.ReportEvent(utility.EVT_SERVER_CLOSE, "Exit Server", nil)
	}
}
