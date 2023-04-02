package controller

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/robinmin/xally/config"
	"github.com/robinmin/xally/shared/model"
	"github.com/robinmin/xally/shared/serverdb"
	"github.com/robinmin/xally/shared/utility"
	log "github.com/sirupsen/logrus"
)

func (h *APIHandler) registerUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var user_info model.UserInfo
		if err := ctx.ShouldBindJSON(&user_info); err != nil {
			log.Error(err.Error())
			h.Response(ctx, config.Text("error_invalid_params"), ERR_INVALID_PARAMETERS, nil)
			return
		}

		// check restrict logic here. If restrict domain is empty, allow all domains
		if config.SvrConfig.Server.EmailRestrictDomain != "" {
			parts := strings.Split(user_info.Email, "@")
			if len(parts) != 2 || strings.ToLower(parts[len(parts)-1]) != strings.ToLower(config.SvrConfig.Server.EmailRestrictDomain) {
				h.Response(ctx, config.Text("error_invalid_email_register"), ERR_REGISTER_FAILED, nil)
				return
			}
		}

		// save user info into database
		var user *serverdb.AuthUser
		var err error
		if user, err = serverdb.RegisterUser(&user_info); err != nil {
			log.Error(err.Error())
			h.Response(ctx, config.Text("error_user_register_failed")+" : "+err.Error(), ERR_REGISTER_FAILED, nil)
			return
		}

		// 发送验证邮件到用户邮箱, 邮件内容包含激活链接，链接中包含一个token
		activation_token, err := serverdb.NewActiviationToken(user.ID)
		if err != nil {
			log.Error(err.Error())
			h.Response(ctx, config.Text("error_generate_token_failed"), ERR_TOKEN_GENERATE_FAILED, nil)
			return
		}
		access_token, err := serverdb.NewAccessToken(user.ID)
		if err != nil {
			log.Error(err.Error())
			h.Response(ctx, config.Text("error_generate_token_failed"), ERR_TOKEN_GENERATE_FAILED, nil)
			return
		}

		if config.SvrConfig.Server.DirectEmailNotify {
			request_url := utility.GetBaseURL(config.SvrConfig.Server.ExternalEndpoint) + "user/activate/" + activation_token.Token
			body := config.GetPageActivate(
				request_url,
				config.Text("tips_email_content_activate"),
				config.Text("tips_email_title_activate"),
				config.Text("tips_email_ignore_msg"),
			)
			if err := utility.SendEmail(user.Email, config.Text("tips_email_subject"), body); err != nil {
				log.Error(err.Error())
				h.Response(ctx, config.Text("error_send_email_failed"), ERR_SENDEMAIL_FAILED, nil)
				return
			}
			h.Response(ctx, config.Text("error_user_register_success"), ERR_OK, gin.H{
				"access_token": access_token.Token,
				"Expired_at":   access_token.ExpiredAt,
			})
		} else {
			h.Response(ctx, config.Text("error_user_register_success2"), ERR_OK, gin.H{
				"access_token": access_token.Token,
				"Expired_at":   access_token.ExpiredAt,
			})
		}
	}
}

func (h *APIHandler) VerifyUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.Param("token")
		if token == "" {
			h.Response(ctx, config.Text("error_invalid_params"), ERR_INVALID_PARAMETERS, nil)
			return
		}

		// load User info from activation token
		user_id, err := serverdb.GetUserIDByActivationToken(token)
		if err != nil || user_id == 0 {
			h.Response(ctx, config.Text("error_invalid_token"), ERR_INVALID_TOKEN, nil)
			return
		}

		// activate user by user_id
		if rows, err := serverdb.ActiviateUser(user_id); err != nil {
			log.Error(err.Error())
			h.Response(ctx, config.Text("tips_email_content_activate_ng"), ERR_ACTIVIATE_FAILED, nil)
			return
		} else {
			tips := config.Text("tips_email_ignore_msg")
			if rows == 1 {
				msg := config.Text("tips_email_content_activate_ok")
				if utility.AcceptJSONResponse(ctx) {
					h.Response(ctx, msg+"\n"+tips, ERR_OK, nil)
				} else {
					ctx.Writer.WriteHeader(http.StatusOK)
					ctx.Writer.Write([]byte(config.GetPageActiviated(msg, config.Text("tips_email_title_activate_ok"), tips)))
				}
			} else {
				msg := config.Text("tips_email_content_activate_ng")
				if utility.AcceptJSONResponse(ctx) {
					h.Response(ctx, msg, ERR_ACTIVIATE_FAILED, nil)
				} else {
					ctx.Writer.WriteHeader(http.StatusOK)
					ctx.Writer.Write([]byte(config.GetPageActiviated(msg, config.Text("tips_email_title_activate_ng"), tips)))
				}
			}
		}
	}
}

// func (h *APIHandler) LoginUser() gin.HandlerFunc {
// 	return func(ctx *gin.Context) {
// 		var user serverdb.AuthUser
// 		if err := ctx.ShouldBindJSON(&user); err != nil {
// 			h.Response(ctx, "参数无效", ERR_INVALID_PARAMETERS, nil)
// 			return
// 		}

// 		real_user, err := user.VerifyUser()
// 		if err != nil {
// 			h.Response(ctx, "用户验证失败或用户当前不可用", ERR_INVALID_USER, nil)
// 		} else {
// 			// 生成JWT token
// 			auth_token := buildToken()
// 			token, err := auth_token.GenerateToken(real_user.ID)
// 			if err != nil {
// 				h.Response(ctx, "生成token失败", ERR_GENERATE_TOKEN_FAILED, nil)
// 			} else {
// 				// TODO: add token into HTTP header
// 				h.Response(ctx, "", ERR_OK, gin.H{
// 					"token": token,
// 				})
// 			}
// 		}
// 	}
// }

// func buildToken() *token.Token {
// 	return &token.Token{
// 		ApiSecret:     config.SvrConfig.Server.AppToken,
// 		TokenLifespan: config.SvrConfig.Server.AppTokenLifespan,
// 	}
// }

// func Logout(c *gin.Context) {
// 	// TODO : load user infor by token
// 	user := serverdb.AuthUser{}
// 	user.Logout()
// }

// func (h *APIHandler) AuthMiddleware() gin.HandlerFunc {
// 	return func(ctx *gin.Context) {
// 		// 从请求头中获取token
// 		authHeader := ctx.GetHeader("Authorization")
// 		if authHeader == "" {
// 			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "缺少Authorization头部"})
// 			ctx.Abort()
// 			return
// 		}

// 		// 解析token
// 		auth_token := buildToken()
// 		token_str := strings.TrimPrefix(authHeader, "Bearer ")
// 		user_id, err := auth_token.GetUserIDFromToken(token_str)
// 		if err != nil {
// 			h.Response(ctx, err.Error(), ERR_INVALID_TOKEN, nil)
// 			ctx.Abort()
// 			return
// 		}

// 		if user_info, err := serverdb.GetValidUser(user_id); err != nil {
// 			h.Response(ctx, "无效的用户ID", ERR_INVALID_USER_ID, nil)
// 			ctx.Abort()
// 			return
// 		} else {
// 			ctx.Set("user_info", user_info)
// 			ctx.Next()
// 		}
// 	}
// }

// func (h *APIHandler) GetUserInfo(ctx *gin.Context) {
// 	// 从上下文中获取用户ID
// 	user_info, exists := ctx.Get("user_info")
// 	if !exists {
// 		h.Response(ctx, "无效的用户ID", ERR_INVALID_USER_ID, nil)
// 		return
// 	}

// 	// TODO: 根据实际情况从数据库中获取用户信息

// 	ctx.JSON(http.StatusOK, user_info)
// }
