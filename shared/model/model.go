package model

import (
	"unicode/utf8"

	gpt3 "github.com/sashabaranov/go-openai"
	"gorm.io/gorm"
)

/******************************************************************************
 * 通用结构体
******************************************************************************/
// type ResponseBody map[string]interface{}
type Response struct {
	Msg  string      `json:"msg"`
	Code uint32      `json:"code"`
	Data interface{} `json:"data,omitempty"`
}

type ConversationHistory struct {
	gorm.Model

	// additional fields
	Role     string `gorm:"size:64;not null;" json:"role"`
	Username string `gorm:"size:64;not null;" json:"username"`

	// from request struct
	AIModel string `gorm:"size:64;not null;" json:"ai_model"`
	MsGSize int    `gorm:"not null;" json:"req_msg_size,omitempty"`

	LatestMsgRole    string `gorm:"size:64;not null;" json:"req_latest_msg_role,omitempty"`
	LatestMsgContent string `gorm:"size:1024;not null;" json:"req_latest_msg_content,omitempty"`

	MaxTokens   int     `json:"req_max_tokens,omitempty"`
	Temperature float32 `json:"req_temperature,omitempty"`
	TopP        float32 `json:"req_top_p,omitempty"`
	N           int     `json:"req_n,omitempty"`
	User        string  `json:"req_user,omitempty"`

	// from response struct
	ID               string `gorm:"size:64;not null;" json:"rsp_id"`
	Object           string `gorm:"size:64;not null;" json:"rsp_object"`
	ChoiceSize       int    `gorm:"not null;" json:"rsp_choice_size,omitempty"`
	PromptTokens     int    `json:"rsp_prompt_tokens,omitempty"`
	CompletionTokens int    `json:"rsp_completion_tokens,omitempty"`
	TotalTokens      int    `json:"rsp_total_tokens,omitempty"`

	LatestChoiceRole         string `gorm:"size:64;" json:"latest_choice_role,omitempty"`
	LatestChoiceContent      string `gorm:"size:1024;" json:"latest_choice_content,omitempty"`
	LatestChoiceName         string `gorm:"size:64;" json:"latest_choice_name,omitempty"`
	LatestChoiceFinishReason string `gorm:"size:64;" json:"latest_choice_finish_reason,omitempty"`
}

func (ch *ConversationHistory) LoadRequest(role string, username string, request *gpt3.ChatCompletionRequest) {
	ch.Role = role
	ch.Username = username
	if request == nil {
		return
	}
	ch.AIModel = request.Model
	ch.MsGSize = len(request.Messages)
	if len(request.Messages) > 0 {
		ch.LatestMsgRole = request.Messages[len(request.Messages)-1].Role
		ch.LatestMsgContent = TruncateStr(request.Messages[len(request.Messages)-1].Content, 1024)
	}

	ch.MaxTokens = request.MaxTokens
	ch.Temperature = request.Temperature
	ch.TopP = request.TopP
	ch.N = request.N
	ch.User = request.User
}

func (ch *ConversationHistory) LoadResponse(response *gpt3.ChatCompletionResponse) {
	if response == nil {
		return
	}

	ch.ID = response.ID
	ch.Object = response.Object
	ch.ChoiceSize = len(response.Choices)
	ch.PromptTokens = response.Usage.PromptTokens
	ch.CompletionTokens = response.Usage.CompletionTokens
	ch.TotalTokens = response.Usage.TotalTokens
	if len(response.Choices) > 0 {
		ch.LatestChoiceRole = response.Choices[len(response.Choices)-1].Message.Role
		ch.LatestChoiceContent = TruncateStr(response.Choices[len(response.Choices)-1].Message.Content, 1024)
		ch.LatestChoiceName = response.Choices[len(response.Choices)-1].Message.Name
		ch.LatestChoiceFinishReason = response.Choices[len(response.Choices)-1].FinishReason
	}
}

func TruncateStr(str string, maxLen int) string {
	if utf8.RuneCountInString(str) > maxLen {
		// 如果字符串长度超过最大长度，则截取前maxLen个字符
		return string([]rune(str)[:maxLen])
	}
	return str
}
