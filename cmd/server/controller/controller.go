package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
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
	ERR_INVALID_USER_ID
	ERR_TOKEN_EXPIRED
	ERR_TOKEN_GENERATE_FAILED
	ERR_DATA_PERSISTENCE
	ERR_REGISTER_FAILED
	ERR_SENDEMAIL_FAILED
	ERR_ACTIVIATE_FAILED
	ERR_DEACTIVIATE_FAILED
	ERR_GENERATE_TOKEN_FAILED
	ERR_UNKNOWN_FAILED
)

const EnableProxyLog = true

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
		AvailableUserMap: map[string]serverdb.WhiteListUser{},
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
func (h *APIHandler) Response(ctx *gin.Context, msg string, biz_code ErrorCode, body gin.H) {
	if utility.AcceptJSONResponse(ctx) {
		rsps := &model.Response{
			Msg:  msg,
			Code: uint32(biz_code),
		}
		if body != nil {
			rsps.Data = body
		}
		ctx.JSON(http.StatusOK, rsps)
	} else {
		err_msg := fmt.Sprintf("[%v] : %s", biz_code, msg)
		http.Error(ctx.Writer, err_msg, http.StatusOK)
	}
}

func (h *APIHandler) ResponseRaw(ctx *gin.Context, msg string, biz_code ErrorCode, body gin.H, code int) {
	if utility.AcceptJSONResponse(ctx) {
		rsps := &model.Response{
			Msg:  msg,
			Code: uint32(biz_code),
		}
		if body != nil {
			rsps.Data = body
		}
		ctx.JSON(code, rsps)
	} else {
		err_msg := fmt.Sprintf("[%v] : %s", biz_code, msg)
		http.Error(ctx.Writer, err_msg, code)
	}
}

// RegisterRoutes 注册路由
func (h *APIHandler) RegisterRoutes(router *gin.Engine, routes *[]config.ProxyRoute) {
	// set default processer
	router.NoRoute(h.noRouteHandler())
	router.NoMethod(h.noMethodHandler())
	router.POST("/user/register/", h.registerUser())
	router.GET("/user/activate/:token", h.VerifyUser())

	for _, rt := range *routes {
		// Just logging the mapping.
		log.Info("Mapping ", rt.Name, " | ", rt.Context, " ---> ", rt.Target)

		target, err := url.Parse(rt.Target)
		if err != nil {
			log.Error("Invalid URL: " + err.Error())
		} else {
			router.Any(
				rt.Context,
				h.authMiddleware(),
				h.reverseProxyHandler(target, config.SvrConfig.Server.OpenaiApiKey, config.SvrConfig.Server.OpenaiOrgID),
			)
		}
	}
}

func (h *APIHandler) reverseProxyHandler(target *url.URL, auth_token string, org_id string) gin.HandlerFunc {
	proxy := httputil.NewSingleHostReverseProxy(target)
	return func(ctx *gin.Context) {
		if _, ok := ctx.Get("auth_user"); ok {
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
		} else {
			h.ResponseRaw(ctx, config.Text("error_invalid_access_denied"), ERR_INVALID_TOKEN, nil, http.StatusUnauthorized)
		}
	}
}

// 无法路由
func (h *APIHandler) noRouteHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		msg := config.Text("error_invalid_url")
		log.Error(msg)
		ctx.JSON(http.StatusNotFound, gin.H{"code": "PAGE_NOT_FOUND", "message": msg})
	}
}

// 不支持的HTTP方法
func (h *APIHandler) noMethodHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		msg := config.Text("error_invalid_http_method")
		log.Error(msg)
		ctx.JSON(http.StatusMethodNotAllowed, gin.H{"code": "METHOD_NOT_ALLOWED", "message": msg})
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
func (h *APIHandler) authMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		reqTime := time.Now()

		// if necessary add sentry tracking first
		if len(config.SvrConfig.Server.SentryDSN) > 0 {
			defer utility.CloseSentry()
		}

		// 检查用户是否在白名单中
		access_token := ctx.Request.Header.Get(config.PROXY_TOKEN_NAME)
		if access_token == "" || !h.WhiteList.IsAccessTokenValid(access_token) {
			// response to client
			h.ResponseRaw(ctx, config.Text("error_invalid_access_denied"), ERR_INVALID_TOKEN, nil, http.StatusUnauthorized)
			return
		}
		ctx.Request.Header.Del(config.PROXY_TOKEN_NAME)

		// // add auth_user into context
		// var user_id uint
		// var err error
		// if user_id, err = serverdb.GetUserIDByAccessToken(access_token); err != nil {
		// 	// response to client
		// 	h.ResponseRaw(ctx, config.Text("error_invalid_token"), ERR_INVALID_TOKEN, nil, http.StatusUnauthorized)
		// 	return
		// } else {
		// 	if auth_user, err := serverdb.GetValidUser(user_id); err != nil {
		// 		// response to client
		// 		h.ResponseRaw(ctx, config.Text("error_invalid_token"), ERR_INVALID_TOKEN, nil, http.StatusUnauthorized)
		// 		return
		// 	} else {
		// 		ctx.Set("auth_user", auth_user)
		// 	}
		// }

		auth_user := h.WhiteList.GetUserInfoByToken(access_token)
		if auth_user != nil {
			ctx.Set("auth_user", auth_user)
		} else {
			h.ResponseRaw(ctx, config.Text("error_invalid_token"), ERR_INVALID_TOKEN, nil, http.StatusUnauthorized)
		}

		// copy request body
		reqBody, _ := io.ReadAll(ctx.Request.Body)
		// And now set a new body, which will simulate the same data we read:
		ctx.Request.Body = io.NopCloser(bytes.NewBuffer(reqBody))

		// 处理请求
		blw := &bodyLogWriter{
			ResponseWriter: ctx.Writer,
			body:           &bytes.Buffer{},
		}

		ctx.Writer = blw
		if len(ctx.Errors) <= 0 {
			ctx.Next()
		}

		// 检查是否有错误
		errs := ctx.Errors.ByType(gin.ErrorTypeAny)
		if len(errs) > 0 {
			var err_msg string
			if len(config.SvrConfig.Server.SentryDSN) > 0 {
				for _, e := range errs {
					err_msg = err_msg + "\n" + e.Err.Error()
					utility.CaptureException(fmt.Errorf("%v", e.Err))
				}
				utility.CaptureRequest(ctx.Request)
				utility.ReportEvent(utility.EVT_SERVER_PROXY_FAILED, config.Text("error_request_failed")+" : "+ctx.Request.URL.String(), nil)
			}
			h.ResponseRaw(ctx, err_msg, ERR_UNKNOWN_FAILED, nil, http.StatusBadRequest)
			return
		}

		if len(config.SvrConfig.Server.SentryDSN) > 0 {
			utility.CaptureRequest(ctx.Request)
			utility.ReportEvent(utility.EVT_SERVER_PROXY_FAILED, config.Text("error_request_success")+" : "+ctx.Request.URL.String(), nil)
		}

		// Refresh token and add back the app_token
		if new_access_token, err := h.WhiteList.RefreshToken(access_token); err == nil {
			ctx.Writer.Header().Set(config.PROXY_TOKEN_NAME, new_access_token)
		} else {
			log.Error("Failed to refresh access token on : " + access_token)
		}

		// write log into db
		if EnableProxyLog {
			reqHeaders, _ := json.Marshal(ctx.Request.Header)
			rspHeaders, _ := json.Marshal(blw.Header())
			plog := &serverdb.ProxyLog{
				RemoteAddr:     ctx.ClientIP(),
				UserID:         auth_user.UserID,
				RequestTime:    reqTime,
				RequestMethod:  ctx.Request.Method,
				RequestURL:     ctx.Request.URL.String(),
				RequestHeaders: string(reqHeaders),
				RequestBody:    string(reqBody),

				ResponseStatusCode: blw.Status(),
				ResponseHeaders:    string(rspHeaders),
				ResponseBody:       blw.body.String(),
			}
			plog.RecordRequest()
		}

		// You can also modify it before sending it out
		if _, err := io.Copy(ctx.Writer, blw.body); err != nil {
			log.Error("Failed to send out response: " + err.Error())
		}
	}
}

// log.Printf("Request: %s %s %s\n", ctx.Request.Method, ctx.Request.URL.String(), ctx.Request.Proto)
// log.Printf("Response: %d %s\n", ctx.Writer.Status(), blw.body.String())

// func headersToString(headers http.Header) string {
// 	headersBytes, _ := json.Marshal(headers)
// 	return string(headersBytes)
// }

// func bodyToString(req *http.Request) string {
// 	bodyBytes, _ := ioutil.ReadAll(req.Body)
// 	req.Body.Close()
// 	req.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
// 	return string(bodyBytes)
// }

// func (h *APIHandler) activateUser() gin.HandlerFunc {
// 	return func(ctx *gin.Context) {
// 		token := ctx.Param("token")
// 		ctx.Writer.WriteHeader(http.StatusOK)
// 		ctx.Writer.Write([]byte(config.GetPageActivate("/user/activated/" + token)))
// 	}
// }

// func (h *APIHandler) completeActivation() gin.HandlerFunc {
// 	return func(ctx *gin.Context) {
// 		ctx.Writer.WriteHeader(http.StatusOK)
// 		ctx.Writer.Write([]byte(config.GetPageActiviated()))
// 	}
// }
