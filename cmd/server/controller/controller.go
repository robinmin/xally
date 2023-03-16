package controller

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/robinmin/xally/config"
	"github.com/robinmin/xally/shared/model"
	"github.com/robinmin/xally/shared/serverdb"
	"github.com/robinmin/xally/shared/utility"
)

type ErrorCode uint32

const ERR_OK = 0
const (
	ERR_INVALID_PARAMETERS ErrorCode = iota + 1000
	ERR_INVALID_USER
	ERR_INVALID_TOKEN
	ERR_TOKEN_EXPIRED
	ERR_DATA_PERSISTENCE
)

// APIHandler 是所有API的处理器
type APIHandler struct {
	WhiteList *serverdb.WhiteList
	// TokenService *token.Token
	DB *gorm.DB
}

// // NewAPIHandler 创建APIHandler实例
func NewAPIHandler(
	api_secret string,
	token_lifespan uint32,
	connection_str string,
	verbose bool,
) (*APIHandler, *gin.Engine) {
	h := &APIHandler{}

	if len(connection_str) > 0 {
		h.DB, _ = serverdb.InitServerDB(connection_str, verbose)
	}

	// 初始化白名单
	h.WhiteList = &serverdb.WhiteList{
		AvailableUserMap: map[string]time.Time{},
		Mutex:            &sync.RWMutex{},
	}

	// 异步更新白名单
	interval := config.SvrConfig.Server.WhiteListRefreshInterval
	if interval < 60 {
		interval = 60
	}
	go h.WhiteList.LoadWhiteList(interval)

	return h, gin.Default()
}

// 通用API响应
func (h *APIHandler) Response(ctx *gin.Context, msg string, biz_code ErrorCode, body interface{}) {
	if body == nil {
		ctx.JSON(http.StatusOK, &model.Response{
			Msg:  msg,
			Code: uint32(biz_code),
		})
	} else {
		ctx.JSON(http.StatusOK, &model.Response{
			Msg:  msg,
			Code: uint32(biz_code),
			Data: body,
		})
	}
}

// RegisterRoutes 注册路由
func (h *APIHandler) RegisterRoutes(router *gin.Engine, routes *[]config.ProxyRoute) {
	// set default processer
	router.NoRoute(h.noRouteHandler())
	router.NoMethod(h.noMethodHandler())

	for _, rt := range *routes {
		// Just logging the mapping.
		log.Info("Mapping ", rt.Name, " | ", rt.Context, " ---> ", rt.Target)

		target, err := url.Parse(rt.Target)
		if err != nil {
			log.Error("Invalid URL: " + err.Error())
		} else {
			router.Any(
				rt.Context,
				h.errorHandlerMiddleware(),
				h.reverseProxyHandler(target, config.SvrConfig.Server.OpenaiApiKey, config.SvrConfig.Server.OpenaiOrgID),
			)
		}
	}

}

func (h *APIHandler) reverseProxyHandler(target *url.URL, auth_token string, org_id string) gin.HandlerFunc {
	proxy := httputil.NewSingleHostReverseProxy(target)
	return func(ctx *gin.Context) {
		// 检查用户是否在白名单中
		access_token := ctx.Request.Header.Get(config.PROXY_TOKEN_NAME)
		if access_token == "" || !h.WhiteList.IsAccessTokenValid(access_token) {
			// Add information into table, automatic registration
			if len(access_token) > 0 {
				if _, err := serverdb.RegisterUser(access_token); err != nil {
					log.Error("Failed to register by access token")
				}
			}

			// response to client
			accpt_type := ctx.Request.Header.Get("Accept")
			err_msg := "Access denied"
			if strings.Contains(strings.ToLower(accpt_type), "application/json") {
				http.Error(ctx.Writer, fmt.Sprintf(`{"code" : %d, "msg" : "%s"}`, http.StatusForbidden, err_msg), http.StatusForbidden)
			} else {
				http.Error(ctx.Writer, err_msg, http.StatusForbidden)
			}
			return
		}

		ctx.Request.Header.Del(config.PROXY_TOKEN_NAME)

		// 替换HTTP头
		ctx.Request.Header.Set("Accept", "application/json; charset=utf-8")
		ctx.Request.Header.Set("Content-Type", "application/json; charset=utf-8")

		ctx.Request.Header.Set("X-Forwarded-Host", ctx.Request.Header.Get("Host"))
		ctx.Request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", auth_token))
		if len(org_id) > 0 {
			ctx.Request.Header.Set("OpenAI-Organization", org_id)
		}

		ctx.Request.Host = target.Host

		proxy.ServeHTTP(ctx.Writer, ctx.Request)
	}
}

// 无法路由
func (h *APIHandler) noRouteHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		log.Error("404 : 无法路由")

		ctx.JSON(http.StatusNotFound, gin.H{"code": "PAGE_NOT_FOUND", "message": "404 page not found"})
	}
}

// 不支持的HTTP方法
func (h *APIHandler) noMethodHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		log.Error("405 : 不支持的HTTP方法")

		ctx.JSON(http.StatusMethodNotAllowed, gin.H{"code": "METHOD_NOT_ALLOWED", "message": "405 method not allowed"})
	}
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// 错误处理中间件
func (h *APIHandler) errorHandlerMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if len(config.SvrConfig.Server.SentryDSN) > 0 {
			defer utility.CloseSentry()
		}

		// 处理请求
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: ctx.Writer}
		ctx.Writer = blw
		if len(ctx.Errors) <= 0 {
			ctx.Next()
		}

		// 检查是否有错误
		if len(config.SvrConfig.Server.SentryDSN) > 0 {
			errs := ctx.Errors.ByType(gin.ErrorTypeAny)
			if len(errs) > 0 {
				for _, e := range errs {
					utility.CaptureException(fmt.Errorf("%v", e.Err))
				}
				utility.CaptureRequest(ctx.Request)
				utility.ReportEvent(utility.EVT_SERVER_PROXY_FAILED, "请求处理出错 : "+ctx.Request.URL.String(), nil)
			} else {
				utility.CaptureRequest(ctx.Request)
				utility.ReportEvent(utility.EVT_SERVER_PROXY_FAILED, "请求处理成功 : "+ctx.Request.URL.String(), nil)

				log.Printf("Request: %s %s %s\n", ctx.Request.Method, ctx.Request.URL.String(), ctx.Request.Proto)
				log.Printf("Response: %d %s\n", ctx.Writer.Status(), blw.body.String())
			}
		}
	}
}
