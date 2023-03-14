package controller

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/robinmin/xally/config"
	"github.com/robinmin/xally/shared/model"
	"github.com/robinmin/xally/shared/serverdb"
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
	WhiteList map[string]time.Time
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

	var err error
	if h.WhiteList, err = serverdb.GetAllUsers(); err != nil {
		log.Error("Filed to load white list", err.Error())
	}

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
	for _, rt := range *routes {
		// Just logging the mapping.
		log.Info("Mapping ", rt.Name, " | ", rt.Context, " ---> ", rt.Target)

		target, err := url.Parse(rt.Target)
		if err != nil {
			log.Error("Invalid URL: " + err.Error())
		} else {
			// router.Any(rt.Context+"/{targetPath:.*}", h.reverseProxyHandler(target, "", ""))
			router.Any(rt.Context, h.reverseProxyHandler(target, config.SvrConfig.Server.OpenaiApiKey, config.SvrConfig.Server.OpenaiOrgID))
		}
	}

}

func (h *APIHandler) reverseProxyHandler(target *url.URL, auth_token string, org_id string) gin.HandlerFunc {
	proxy := httputil.NewSingleHostReverseProxy(target)
	return func(ctx *gin.Context) {
		// 检查用户是否在白名单中
		access_token := ctx.Request.Header.Get(config.PROXY_TOKEN_NAME)
		if access_token == "" || !h.validAccessToken(access_token) {
			// Add information into table, automatic registration
			if _, err := serverdb.RegisterUser(access_token); err != nil {
				log.Error("Failed to register by access token")
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

func (h *APIHandler) validAccessToken(access_token string) bool {
	if expiry_date, ok := h.WhiteList[access_token]; !ok || expiry_date.Local().Before(time.Now()) {
		log.Error("Token is invalid or already expired!")
		return false
	}
	return true
}
