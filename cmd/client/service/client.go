package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path"

	"github.com/robinmin/xally/cmd/server/controller"
	"github.com/robinmin/xally/config"
	"github.com/robinmin/xally/shared/model"
	"github.com/robinmin/xally/shared/utility"
	gpt3 "github.com/sashabaranov/go-openai"
	log "github.com/sirupsen/logrus"
)

type ChatGPTCLient struct {
	gpt3.Client
	HTTPClient *http.Client
}

// CreateChatCompletion — API call to Create a completion for the chat message.
func (c *ChatGPTCLient) CreateChatCompletion(
	ctx context.Context,
	request gpt3.ChatCompletionRequest,
) (response gpt3.ChatCompletionResponse, err error) {
	model := request.Model
	if model != gpt3.GPT3Dot5Turbo0301 && model != gpt3.GPT3Dot5Turbo && model != gpt3.GPT3TextDavinci003 && model != gpt3.GPT3TextDavinci002 && model != gpt3.GPT3TextCurie001 && model != gpt3.GPT3TextBabbage001 && model != gpt3.GPT3TextAda001 && model != gpt3.GPT3TextDavinci001 && model != gpt3.GPT3Davinci && model != gpt3.GPT3Curie && model != gpt3.GPT3Ada && model != gpt3.GPT3Babbage && model != gpt3.CodexCodeDavinci002 && model != gpt3.CodexCodeCushman001 && model != gpt3.CodexCodeDavinci001 {
		err = gpt3.ErrChatCompletionInvalidModel
		return
	}

	var reqBytes []byte
	reqBytes, err = json.Marshal(request)
	if err != nil {
		return
	}

	urlSuffix := "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.fullURL(urlSuffix), bytes.NewBuffer(reqBytes))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
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
