package model

import (
	"os"
	"os/user"
	"time"

	"github.com/denisbrodbeck/machineid"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

/******************************************************************************
 * 通用结构体
******************************************************************************/
// type ResponseBody map[string]interface{}
type Response struct {
	Msg  string `json:"msg"`
	Code uint32 `json:"code"`
	Data gin.H  `json:"data,omitempty"`
}

type ConversationHistory struct {
	// gorm.Model
	ID        string `gorm:"type:char(36); primaryKey; not null;"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// additional fields
	Role     string `gorm:"type:varchar(64); not null;" json:"role"`
	Username string `gorm:"type:varchar(64); not null;" json:"username"`

	// from request struct
	AIModel string `gorm:"type:varchar(64); not null;" json:"ai_model"`
	MsGSize int    `gorm:"not null; default:0;" json:"req_msg_size,omitempty"`

	LatestMsgRole    string `gorm:"type:varchar(64); not null;" json:"req_latest_msg_role,omitempty"`
	LatestMsgContent string `gorm:"type:text;" json:"req_latest_msg_content,omitempty"`

	MaxTokens   int     `json:"req_max_tokens,omitempty"`
	Temperature float32 `json:"req_temperature,omitempty"`
	TopP        float32 `json:"req_top_p,omitempty"`
	N           int     `json:"req_n,omitempty"`
	User        string  `gorm:"type:varchar(64);" json:"req_user,omitempty"`

	// from response struct
	ResponseID       string `gorm:"type:varchar(64); not null;" json:"response_id"`
	Object           string `gorm:"type:varchar(64); not null;" json:"rsp_object"`
	ChoiceSize       int    `gorm:"not null; default:0;" json:"rsp_choice_size,omitempty"`
	PromptTokens     int    `json:"rsp_prompt_tokens,omitempty"`
	CompletionTokens int    `json:"rsp_completion_tokens,omitempty"`
	TotalTokens      int    `json:"rsp_total_tokens,omitempty"`

	LatestChoiceRole         string `gorm:"type:varchar(64);" json:"latest_choice_role,omitempty"`
	LatestChoiceContent      string `gorm:"type:text;" json:"latest_choice_content,omitempty"`
	LatestChoiceName         string `gorm:"type:varchar(64);" json:"latest_choice_name,omitempty"`
	LatestChoiceFinishReason string `gorm:"type:varchar(256);" json:"latest_choice_finish_reason,omitempty"`
}

type UserInfo struct {
	Username   string    `json:"username"`
	UserID     string    `json:"userid"`
	Hostname   string    `json:"hostname"`
	Email      string    `json:"email"`
	DeviceInfo string    `json:"device_info"`
	Password   string    `json:"password"`
	AppToken   string    `json:"app_token"`
	RegisterAt time.Time `json:"register_at"`
}

func NewUserInfo(app_token string, email string, password string) (*UserInfo, error) {
	var err error
	var device_info string
	var current_user *user.User

	// get user id
	current_user, err = user.Current()
	if err != nil {
		return nil, err
	}

	// get device info
	if len(app_token) > 0 {
		device_info, err = machineid.ProtectedID(app_token)
	} else {
		device_info, err = machineid.ID()
	}
	if err != nil {
		return nil, err
	}

	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}

	userform := &UserInfo{
		Username:   current_user.Username,
		UserID:     current_user.Uid,
		Hostname:   hostname,
		Email:      email,
		DeviceInfo: device_info,
		Password:   password,
		AppToken:   app_token,
		RegisterAt: time.Now(),
	}
	return userform, nil
}
