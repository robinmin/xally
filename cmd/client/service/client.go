package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/robinmin/xally/config"
	"github.com/robinmin/xally/shared/utility"
	gpt3 "github.com/sashabaranov/go-openai"
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
	if model != gpt3.GPT3Dot5Turbo0301 && model != gpt3.GPT3Dot5Turbo {
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

func (c *ChatGPTCLient) sendRequest(req *http.Request, v interface{}) error {
	///////////////////////////////////////////////////////////////////////////
	// Add user defined header here
	if config.MyConfig.IsSharedMode() && len(config.MyConfig.System.Email) > 0 {
		if token, err := utility.GenerateAccessToken(config.MyConfig.System.SharedToken, config.MyConfig.System.Email); err == nil {
			req.Header.Set(config.PROXY_TOKEN_NAME, token)
		}
	}
	///////////////////////////////////////////////////////////////////////////

	req.Header.Set("Accept", "application/json; charset=utf-8")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.MyConfig.System.APIKeyOpenai))

	// Check whether Content-Type is already set, Upload Files API requires
	// Content-Type == multipart/form-data
	contentType := req.Header.Get("Content-Type")
	if contentType == "" {
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
	}

	if len(config.MyConfig.System.APIOrgIDOpenai) > 0 {
		req.Header.Set("OpenAI-Organization", config.MyConfig.System.APIOrgIDOpenai)
	}

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		var errRes gpt3.ErrorResponse
		err = json.NewDecoder(res.Body).Decode(&errRes)
		if err != nil || errRes.Error == nil {
			reqErr := gpt3.RequestError{
				StatusCode: res.StatusCode,
				Err:        err,
			}
			return fmt.Errorf("error, %w", &reqErr)
		}
		errRes.Error.StatusCode = res.StatusCode
		return fmt.Errorf("error, status code: %d, message: %w", res.StatusCode, errRes.Error)
	}

	if v != nil {
		if err = json.NewDecoder(res.Body).Decode(v); err != nil {
			return err
		}
	}

	return nil
}
