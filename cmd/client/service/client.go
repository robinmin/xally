package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path"
	"strconv"

	"github.com/google/uuid"
	gpt3 "github.com/sashabaranov/go-openai"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"

	"github.com/robinmin/xally/cmd/server/controller"
	"github.com/robinmin/xally/config"
	"github.com/robinmin/xally/shared/model"
	"github.com/robinmin/xally/shared/utility"
)

type ChatGPTCLient struct {
	gpt3.Client

	HTTPClient     *http.Client
	msg_history    []gpt3.ChatCompletionMessage
	support_models map[string]ModelData
}

type ObjectType string

const (
	OTModel           ObjectType = "model"
	OTModelPermission ObjectType = "model_permission"
	OTList            ObjectType = "list"
	OTEdit            ObjectType = "edit"
	OTTextCompletion  ObjectType = "text_completion"
	OTEEmbedding      ObjectType = "embedding"
	OTFile            ObjectType = "file"
	OTFineTune        ObjectType = "fine-tune"
	OTFineTuneEvent   ObjectType = "fine-tune-event"
)

type ModelData struct {
	ID         string            `json:"id"`
	Object     ObjectType        `json:"object"`
	Created    int64             `json:"created"`
	OwnedBy    string            `json:"owned_by"`
	Permission []ModelPermission `json:"permission"`
	Root       string            `json:"root"`
	Parent     string            `json:"parent"`
}

type ModelPermission struct {
	ID                 string     `json:"id"`
	Object             ObjectType `json:"object"`
	Created            int64      `json:"created"`
	AllowCreateEngine  bool       `json:"allow_create_engine"`
	AllowSampling      bool       `json:"allow_sampling"`
	AllowLogProbs      bool       `json:"allow_logprobs"`
	AllowSearchIndices bool       `json:"allow_search_indices"`
	AllowView          bool       `json:"allow_view"`
	AllowFineTuning    bool       `json:"allow_fine_tuning"`
	Organization       string     `json:"organization"`
	Group              string     `json:"group"`
	IsBlocking         bool       `json:"is_blocking"`
}

type ModelsListResponse struct {
	Data   []ModelData `json:"data"`
	Object ObjectType
}

func NewChatBotClient(api_key string, api_endpoint string) *ChatGPTCLient {
	// build the client object with existing API keys and API endpoints
	api_cfg := gpt3.DefaultConfig(api_key)
	if len(api_endpoint) > 0 {
		api_cfg.BaseURL = api_endpoint
	}

	log.Println("api_cfg.BaseURL  = ", api_cfg.BaseURL)
	client := &ChatGPTCLient{
		Client:         *gpt3.NewClientWithConfig(api_cfg),
		HTTPClient:     api_cfg.HTTPClient,
		support_models: map[string]ModelData{},
	}
	client.ResetMsgHistory("system", "")

	return client
}

// CreateChatCompletionEx — API call to Create a completion for the chat message.
func (c *ChatGPTCLient) CreateChatCompletionEx(
	model string,
	token_len int,
	temperature float32,
	role_name string,
	username string,
	chat_history *model.ConversationHistory,
	// ctx context.Context,
	// request gpt3.ChatCompletionRequest,
) (response gpt3.ChatCompletionResponse, err error) {
	// current_model := bot.role.Model
	if model == "" {
		model = gpt3.GPT3Dot5Turbo
	}
	request := gpt3.ChatCompletionRequest{
		Model:     model,
		Messages:  c.msg_history,
		MaxTokens: token_len,
		// MaxTokens:   3000,
		Temperature: temperature,
		N:           1,
	}

	// fetch support models at the first time
	c.getSupportModels()

	if _, ok := c.support_models[model]; !ok {
		err = gpt3.ErrChatCompletionInvalidModel
		return
	}

	var reqBytes []byte
	reqBytes, err = json.Marshal(request)
	if err != nil {
		return
	}

	urlSuffix := "/chat/completions"
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, c.fullURL(urlSuffix), bytes.NewBuffer(reqBytes))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)

	// chat_history = &model.ConversationHistory{}
	loadRequest(chat_history, role_name, username, &request)
	loadResponse(chat_history, &response)

	return
}

func (c *ChatGPTCLient) getSupportModels() error {
	if len(c.support_models) > 0 {
		return nil
	}

	// fetch available models
	hdr := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", config.MyConfig.System.OpenaiApiKey),
	}

	resp_code, resp_body, err := utility.FetchURL("GET", c.fullURL("/models"), "", hdr)
	if resp_code == http.StatusOK {
		var list_models ModelsListResponse
		err = json.Unmarshal([]byte(resp_body), &list_models)
		if err != nil {
			log.Error("Failed to unmarshal HTTP response body: " + err.Error())
		} else {
			for _, model := range list_models.Data {
				log.Debug(model.Object, " : ", model.ID)
				c.support_models[model.ID] = model
			}
		}
	} else {
		log.Error("Failed to list all models")
		err = errors.New("Failed to list all models, invalid response code: " + strconv.Itoa(resp_code))
	}
	return err
}

func (c *ChatGPTCLient) fullURL(suffix string) string {
	return fmt.Sprintf("%s%s", config.MyConfig.System.APIEndpointOpenai, suffix)
}

func (c *ChatGPTCLient) sendRequest(req *http.Request, val interface{}) error {
	///////////////////////////////////////////////////////////////////////////
	// Add user defined headers here
	if config.MyConfig.IsSharedMode() {
		if len(config.MyConfig.System.AppToken) > 0 {
			// if tk, err := token.GenerateAccessToken(config.MyConfig.System.AppToken, config.MyConfig.System.Email); err == nil {
			req.Header.Set(config.PROXY_TOKEN_NAME, config.MyConfig.System.AppToken)
			// }
		} else {
			return errors.New("Please set APP_TOKEN and Email in sharing mode")
		}
	} else {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.MyConfig.System.OpenaiApiKey))
		if len(config.MyConfig.System.OpenaiOrgID) > 0 {
			req.Header.Set("OpenAI-Organization", config.MyConfig.System.OpenaiOrgID)
		}
	}
	///////////////////////////////////////////////////////////////////////////

	acceptType := req.Header.Get("Accept")
	if acceptType == "" {
		req.Header.Set("Accept", "application/json; charset=utf-8")
	}
	acceptLang := req.Header.Get("Accept-Language")
	if acceptLang == "" {
		req.Header.Set("Accept-Language", "application/json; charset=utf-8")
	}
	// Check whether Content-Type is already set, Upload Files API requires
	// Content-Type == multipart/form-data
	contentType := req.Header.Get("Content-Type")
	if contentType == "" {
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
	}

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	// filter out internal errors
	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		var errRes model.Response
		err = json.NewDecoder(res.Body).Decode(&errRes)
		if err != nil {
			return fmt.Errorf("error, status code: %d, message: invalid response body.", res.StatusCode)
		}

		return fmt.Errorf(
			"error, http status code: %d, biz_code: %d, message: %s",
			res.StatusCode,
			errRes.Code,
			errRes.Msg,
		)
	}

	// copy response data
	if val != nil {
		if err = json.NewDecoder(res.Body).Decode(val); err != nil {
			return err
		}
	}

	// update refreshed token if necessary and update into YAML file
	if config.MyConfig.IsSharedMode() {
		new_access_token := res.Header.Get(config.PROXY_TOKEN_NAME)
		if len(new_access_token) > 0 && new_access_token != config.MyConfig.System.AppToken {
			config.MyConfig.System.AppToken = new_access_token

			var temp_file string
			var dir_home string
			var err error

			if dir_home, err = config.FindHomeDir(false); err != nil {
				dir_home = "."
			}
			temp_file = path.Join(dir_home, "xally.yaml")
			if _, err = config.MyConfig.DumpIntoYAML(temp_file); err != nil {
				log.Error("Failed to write YAML data into :" + temp_file)
			}
		}
	}

	return nil
}

func (c *ChatGPTCLient) UserLogin() (string, error) {
	return "", nil
}

func (c *ChatGPTCLient) UserLogout() (string, error) {
	return "", nil
}

func (c *ChatGPTCLient) UserRegistration(email string, endpoint_url string) (string, error) {
	var msg string

	if !utility.IsValidEmail(email) {
		msg = fmt.Sprintf(config.Text("error_invalid_email"), email)
		return msg, nil
	}

	if !utility.IsValidURL(endpoint_url) {
		msg = fmt.Sprintf(config.Text("error_invalid_endpoint_url"), endpoint_url)
		return msg, nil
	}

	// get remote URL
	request, err := model.NewUserInfo("", email, "")
	if err != nil {
		return "", err
	}

	var reqBytes []byte
	reqBytes, err = json.Marshal(request)
	if err != nil {
		return err.Error(), err
	}

	request_url := utility.GetBaseURL(endpoint_url) + "user/register/"
	code, resp_body, err := utility.FetchURL("POST", request_url, string(reqBytes), nil)
	if code != http.StatusOK || err != nil {
		if err == nil {
			msg = fmt.Sprintf("error, status code: %d, body: %s", code, resp_body)
		} else {
			msg = fmt.Sprintf("error, status code: %d, message: %s, body: %s", code, err.Error(), resp_body)
		}
		return msg, err
	}

	if len(resp_body) != 0 {
		var result model.Response
		err = json.Unmarshal([]byte(resp_body), &result)
		if err != nil {
			log.Error("Failed to unmarshal HTTP response body: " + err.Error())
		} else {
			if result.Code == controller.ERR_OK && result.Data != nil {
				access_token, ok := result.Data["access_token"].(string)
				if ok {
					var temp_file string
					var dir_home string

					// update local configuration
					config.MyConfig.System.Email = email
					config.MyConfig.System.APIEndpointOpenai = endpoint_url
					config.MyConfig.System.UseSharedMode = 1
					config.MyConfig.System.AppToken = access_token

					var err error
					if dir_home, err = config.FindHomeDir(false); err != nil {
						dir_home = "."
					}
					temp_file = path.Join(dir_home, "xally.yaml")
					if _, err = config.MyConfig.DumpIntoYAML(temp_file); err != nil {
						log.Error("Failed to write YAML data into :" + temp_file)
					}
					return result.Msg, nil
				}
			} else {
				log.Errorf("Failed to register user [%v] : %s ", result.Code, result.Msg)
			}
		}
	}
	return resp_body, nil
}

// please refer to this page for limitation:
//
//	https://platform.openai.com/docs/models/gpt-3-5
func (c *ChatGPTCLient) GetMaxTokens(model string) int {
	switch model {
	case gpt3.GPT3Dot5Turbo0301:
		return 4096
	case gpt3.GPT3Dot5Turbo:
		return 4096
	case gpt3.GPT3TextDavinci003:
		return 4097
	case gpt3.GPT3TextDavinci002:
		return 4097
	case gpt3.GPT3TextCurie001:
		return 2049
	case gpt3.GPT3TextBabbage001:
		return 2049
	case gpt3.GPT3TextAda001:
		return 2049
	case gpt3.GPT3TextDavinci001:
		return 2049
	// case gpt3.GPT3DavinciInstructBeta:
	// 	return 2049
	case gpt3.GPT3Davinci:
		return 2049
	// case gpt3.GPT3CurieInstructBeta:
	// 	return 2049
	case gpt3.GPT3Curie:
		return 2049
	case gpt3.GPT3Ada:
		return 2049
	case gpt3.GPT3Babbage:
		return 2049
	case gpt3.CodexCodeDavinci002:
		return 8001
	case gpt3.CodexCodeCushman001:
		return 2048
	case gpt3.CodexCodeDavinci001:
		return 8001
	default:
		return 4096
	}
}

func (c *ChatGPTCLient) ResetMsgHistory(prompt string, opening string) {
	// refresh role relevant variables
	c.msg_history = []gpt3.ChatCompletionMessage{
		{
			Role:    "system",
			Content: prompt,
		},
	}
	if len(opening) > 0 {
		c.AddMsgHistory("user", opening)
	}
}

func (c *ChatGPTCLient) AddMsgHistory(role_name string, content string) {
	c.msg_history = append(c.msg_history, gpt3.ChatCompletionMessage{
		Role:    role_name,
		Content: content,
	})
}

func (c *ChatGPTCLient) EstimateAvailableTokenNumber(model string, question_leng int) int {
	available_len := c.GetMaxTokens(model) - question_leng // - bot.token_counter_total

	for _, msg := range c.msg_history {
		available_len = available_len - 4 // every message follows <im_start>{role/name}\n{content}<im_end>\n
		// TODO: need to fine tune going forward, replaced with GPT-index instead of length of contents
		available_len = available_len - len(msg.Content) - len(msg.Name)
		available_len = available_len - 1 // - len(msg.Role), role is always required and always 1 token
		available_len = available_len - 2 // every reply is primed with <im_start>assistant
	}

	return available_len
}

func (c *ChatGPTCLient) AdjustMsgHistory(init_msg_len int, question_len int, model string) (bool, int) {
	var token_len int
	need_reset := false

	// adjust available max token length
	for {
		token_len = c.EstimateAvailableTokenNumber(model, question_len)
		if token_len > 0 {
			break
		}

		if len(c.msg_history) < init_msg_len {
			// we assume the default configuration is fine
			// bot.resetRole(bot.role.Name, true)
			need_reset = true
			token_len = c.EstimateAvailableTokenNumber(model, question_len)

			if config.MyConfig.DebugMode() {
				fmt.Println("🌸🌸🌸🌸🌸🌸🌸🌸🌸🌸🌸🌸🌸🌸🌸🌸 resetting role......")
			}
			break
		} else {
			// popup the old conversation history but key the prompt and the potential opening
			size_before := len(c.msg_history)
			c.msg_history = slices.Delete(c.msg_history, init_msg_len, init_msg_len+1)

			if config.MyConfig.DebugMode() {
				fmt.Print("🌻🌻🌻🌻🌻🌻🌻🌻🌻🌻🌻🌻🌻🌻🌻🌻 erasing old history......")
				fmt.Printf("%d --> %d\n", size_before, len(c.msg_history))
			}
		}
	}
	return need_reset, token_len
}

func (c *ChatGPTCLient) GetMsgHistoryLength() int {
	return len(c.msg_history)
}

func loadRequest(ch *model.ConversationHistory, role string, username string, request *gpt3.ChatCompletionRequest) {
	ch.ID = uuid.New().String()
	ch.Role = role
	ch.Username = username
	if request == nil {
		return
	}
	ch.AIModel = request.Model
	ch.MsGSize = len(request.Messages)
	if len(request.Messages) > 0 {
		ch.LatestMsgRole = request.Messages[len(request.Messages)-1].Role
		ch.LatestMsgContent = request.Messages[len(request.Messages)-1].Content
	}

	ch.MaxTokens = request.MaxTokens
	ch.Temperature = request.Temperature
	ch.TopP = request.TopP
	ch.N = request.N
	ch.User = request.User
}

func loadResponse(ch *model.ConversationHistory, response *gpt3.ChatCompletionResponse) {
	if response == nil {
		return
	}

	ch.ResponseID = response.ID
	ch.Object = response.Object
	ch.ChoiceSize = len(response.Choices)
	ch.PromptTokens = response.Usage.PromptTokens
	ch.CompletionTokens = response.Usage.CompletionTokens
	ch.TotalTokens = response.Usage.TotalTokens
	if len(response.Choices) > 0 {
		ch.LatestChoiceRole = response.Choices[len(response.Choices)-1].Message.Role
		ch.LatestChoiceContent = response.Choices[len(response.Choices)-1].Message.Content
		ch.LatestChoiceName = response.Choices[len(response.Choices)-1].Message.Name
		ch.LatestChoiceFinishReason = response.Choices[len(response.Choices)-1].FinishReason
	}
}
