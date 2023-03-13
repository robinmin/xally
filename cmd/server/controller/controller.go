package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/robinmin/xally/shared/serverdb"
	"github.com/robinmin/xally/shared/utility/token"
	"gorm.io/gorm"
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
	TokenService *token.Token
	DB           *gorm.DB
}

// // NewAPIHandler 创建APIHandler实例
func NewAPIHandler(api_secret string, token_lifespan uint32, connection_str string, verbose bool) (*APIHandler, *gin.Engine) {
	var db *gorm.DB
	if len(connection_str) > 0 {
		db, _ = serverdb.InitServerDB(connection_str, verbose)
	}
	return &APIHandler{
		TokenService: &token.Token{ApiSecret: api_secret, TokenLifespan: token_lifespan},
		DB:           db,
	}, gin.Default()
}

// RegisterRoutes 注册路由
func (h *APIHandler) RegisterRoutes(router *gin.Engine) {
	// router.GET("/", h.defaultHandler())
	// router.POST("/", h.defaultHandler())
	// router.GET(model.URL_SYS_INDEX, h.defaultHandler())
	// router.POST(model.URL_SYS_INDEX, h.defaultHandler())

	// 	// default routes for error handling
	// 	router.NoRoute(h.noRouteHandler())
	// 	router.NoMethod(h.noMethodHandler())

	// 	api_router := router.Group(model.URL_PREFIX_API)
	// 	{
	// 		// 用户注册
	// 		api_router.POST(model.URL_USER_REGISTER, h.registerHandler())

	// 		// 用户激活
	// 		api_router.POST(model.URL_USER_ACTIVATE, h.activateHandler())

	// 		// 用户注销
	// 		api_router.POST(model.URL_USER_DEACTIVATE, h.deactivateHandler())

	// 		// 用户查询
	// 		api_router.POST(model.URL_USER_QUERY, h.JwtAuthMiddleware(), h.askHandler())

	// 		// 登录
	// 		api_router.POST(model.URL_USER_LOGIN, h.loginHandler())

	// 		// 登出
	// 		api_router.POST(model.URL_USER_LOGOUT, h.logoutHandler())
	// 	}

	// 	page_router := router.Group(model.URL_PREFIX_PAGE)
	// 	{
	// 		// 主页
	// 		page_router.GET(model.URL_SYS_HOME, h.homeHandler())

	// }
}

// // 通用API响应
// func (h *APIHandler) Response(ctx *gin.Context, msg string, biz_code ErrorCode, body interface{}) {
// 	if body == nil {
// 		ctx.JSON(http.StatusOK, &model.Response{
// 			Msg:  msg,
// 			Code: uint32(biz_code),
// 		})
// 	} else {
// 		ctx.JSON(http.StatusOK, &model.Response{
// 			Msg:  msg,
// 			Code: uint32(biz_code),
// 			Data: body,
// 		})
// 	}
// }

// // 处理用户注册
// func (h *APIHandler) registerHandler() gin.HandlerFunc {
// 	return func(ctx *gin.Context) {
// 		var request model.UserRegisterRequest

// 		log.Debug("User register starting......")
// 		err := ctx.ShouldBindJSON(&request)
// 		if err != nil {
// 			log.Debug("User register failed")
// 			h.Response(ctx, err.Error(), ERR_INVALID_PARAMETERS, nil)
// 			return
// 		}
// 		log.Debug(request)

// 		// user registration logic here
// 		user := &serverdb.User{
// 			Username:   request.Username,
// 			Password:   request.Password,
// 			Email:      request.Email,
// 			DeviceInfo: request.DeviceInfo,
// 			Enabled:    0, // default to 0, when be enabled after activated
// 			RegisterAt: time.Now(),
// 		}

// 		rows, err := user.SaveUser()
// 		if err != nil {
// 			h.Response(ctx, err.Error(), ERR_INVALID_PARAMETERS, nil)
// 			return
// 		}
// 		if rows == 1 {
// 			log.Debug("User register success")

// 			h.Response(ctx, "", ERR_OK, &model.UserRegisterResponseBody{
// 				Username: user.Username,
// 				Token:    "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
// 			})
// 		} else {
// 			h.Response(ctx, "No user has been registered.", ERR_DATA_PERSISTENCE, model.Response{
// 				Msg:  "",
// 				Code: 0,
// 			})
// 		}
// 	}
// }

// // 处理用户激活
// func (h *APIHandler) activateHandler() gin.HandlerFunc {
// 	return func(ctx *gin.Context) {
// 		log.Error("TODO: 处理用户激活逻辑")
// 	}
// }

// // 处理用户注销
// func (h *APIHandler) deactivateHandler() gin.HandlerFunc {
// 	return func(ctx *gin.Context) {
// 		log.Error("TODO: 处理用户注销逻辑")
// 	}
// }

// // 处理用户查询
// func (h *APIHandler) askHandler() gin.HandlerFunc {
// 	return func(ctx *gin.Context) {
// 		log.Error("TODO: 处理用户查询逻辑")
// 	}
// }

// // 处理登录
// func (h *APIHandler) loginHandler() gin.HandlerFunc {
// 	return func(ctx *gin.Context) {
// 		var request model.UserLoginRequest

// 		log.Debug("User login starting......")
// 		err := ctx.ShouldBindJSON(&request)
// 		if err != nil {
// 			log.Debug("User login failed")
// 			h.Response(ctx, err.Error(), ERR_INVALID_PARAMETERS, nil)
// 			return
// 		}
// 		log.Debug(request)

// 		// if appID != "" {
// 		// 	id, err = machineid.ProtectedID(appID)
// 		// } else {
// 		// 	id, err = machineid.ID()
// 		// }
// 		// TODO: find out user information by username and device information

// 		// user login logic here
// 		user := serverdb.User{}
// 		user.Username = request.Username
// 		// user.Password = request.Password
// 		token, err := serverdb.LoginCheck(user.Username, user.Password)
// 		if err != nil {
// 			h.Response(ctx, err.Error(), ERR_INVALID_TOKEN, nil)
// 			return
// 		}

// 		log.Debug("User login success")
// 		h.Response(ctx, "", ERR_OK, &model.UserLoginResponseBody{
// 			Token: token,
// 		})
// 	}
// }

// // 处理登出
// func (h *APIHandler) logoutHandler() gin.HandlerFunc {
// 	return func(ctx *gin.Context) {
// 		log.Error("TODO: 处理登出逻辑")
// 	}
// }

// // 处理系统主页面
// func (h *APIHandler) homeHandler() gin.HandlerFunc {
// 	return func(ctx *gin.Context) {
// 		log.Error("TODO: 处理系统主页面逻辑")
// 	}
// }

// // 默认首页
// func (h *APIHandler) defaultHandler() gin.HandlerFunc {
// 	return func(ctx *gin.Context) {
// 		log.Error("默认首页")
// 	}
// }

// // 无法路由
// func (h *APIHandler) noRouteHandler() gin.HandlerFunc {
// 	return func(ctx *gin.Context) {
// 		log.Error("404 : 无法路由")

// 		ctx.JSON(http.StatusNotFound, gin.H{"code": "PAGE_NOT_FOUND", "message": "404 page not found"})
// 	}
// }

// // 不支持的HTTP方法
// func (h *APIHandler) noMethodHandler() gin.HandlerFunc {
// 	return func(ctx *gin.Context) {
// 		log.Error("405 : 不支持的HTTP方法")

// 		ctx.JSON(http.StatusMethodNotAllowed, gin.H{"code": "METHOD_NOT_ALLOWED", "message": "405 method not allowed"})
// 	}
// }

// // 错误处理中间件
// func (h *APIHandler) errorHandlerMiddleware() gin.HandlerFunc {
// 	return func(ctx *gin.Context) {
// 		log.Error("错误处理中间件")
// 		log.Printf("Total Errors -> %d", len(ctx.Errors))

// 		if len(ctx.Errors) <= 0 {
// 			ctx.Next()
// 			return
// 		}

// 		for _, err := range ctx.Errors {
// 			log.Printf("Error -> %+v\n", err)
// 		}
// 		ctx.JSON(http.StatusInternalServerError, "")
// 	}
// }

// // // 鉴权中间件，用于保护系统主页面
// // func (h *APIHandler) authMiddleware() gin.HandlerFunc {
// // 	return func(ctx *gin.Context) {
// // 		log.Error("TODO: 鉴权逻辑")
// // 	}
// // }

// func (h *APIHandler) JwtAuthMiddleware() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		err := h.CurrentToken.TokenValid(c)
// 		if err != nil {
// 			c.String(http.StatusUnauthorized, "Unauthorized")
// 			c.Abort()
// 			return
// 		}
// 		c.Next()
// 	}
// }
