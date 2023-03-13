package chatbot

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/c-bata/go-prompt"
	"github.com/fatih/color"
	strftime "github.com/itchyny/timefmt-go"
	gpt3 "github.com/sashabaranov/go-gpt3"
	log "github.com/sirupsen/logrus"

	"robinmin.net/tools/xally/config"
	"robinmin.net/tools/xally/shared/utility"
)

const default_user_avatar = "👦"

type suggestionType int

const (
	// execute host command on locase machine
	Ask suggestionType = iota

	// translate text
	Translate

	// execute host command on locase machine
	Cmd

	// Reset suggestion
	Reset

	// Quit suggestion
	Quit

	// Others is key for various arbitrary suggestions
	Others
)

const prompt_tip_flag = " ▶ "

func get_suggestion_map(role_name string) *map[suggestionType][]prompt.Suggest {
	reset_rolese := []prompt.Suggest{}
	for _, tmp_role := range config.MyConfig.Roles {
		reset_rolese = append(reset_rolese, prompt.Suggest{
			Text:        "reset " + tmp_role.Name,
			Description: config.Text("tips_suggestion_reset") + tmp_role.Name + tmp_role.Avatar,
		})
	}

	suggestionsMap := &map[suggestionType][]prompt.Suggest{
		Ask: {
			{Text: "ask", Description: config.Text("tips_suggestion_ask")},
		},
		Translate: {
			{Text: "translate", Description: config.Text("tips_suggestion_translate")},
			{Text: "lookup", Description: config.Text("tips_suggestion_translate")},
		},
		Cmd: {
			{Text: "cmd", Description: config.Text("tips_suggestion_cmd")},
		},
		Reset: reset_rolese,
		Quit: {
			{Text: "q、88", Description: config.Text("tips_suggestion_quit")},
			// {Text: "88", Description: config.Text("tips_suggestion_quit")},
			// {Text: "886", Description: config.Text("tips_suggestion_quit")},
			// {Text: "bye", Description: config.Text("tips_suggestion_quit")},
			// {Text: "quit", Description: config.Text("tips_suggestion_quit")},
			{Text: "exit、quit", Description: config.Text("tips_suggestion_quit")},
		},
		Others: {},
	}
	return suggestionsMap
}

type ChatBot struct {
	// chatbot name
	name string
	role *config.SysRole

	// logfile for chat conversation
	chat_history_file *os.File
	chat_history_path string

	client        *gpt3.Client
	msg_history   []gpt3.ChatCompletionMessage
	log_history   bool
	token_counter int
	prompt        *prompt.Prompt
}

func NewChatbot(chat_history_path string, name string, role_name string, log_history bool) *ChatBot {
	bot := &ChatBot{
		name:              name,
		chat_history_file: nil,
		chat_history_path: chat_history_path,
		client:            nil,
		msg_history:       nil,
		log_history:       log_history,
		token_counter:     0,
		prompt:            nil,
	}
	if role_name == "" {
		bot.resetRole(config.MyConfig.System.DefaultRole, true)
	} else {
		bot.resetRole(role_name, true)
	}

	api_key := config.MyConfig.System.APIKeyOpenai
	if api_key == "" {
		api_key = os.Getenv("OPENAI_API_KEY")
		if api_key == "" {
			bot.Say("- "+config.Text("error_no_chatgpt_key"), true)
			return bot
		}
	}

	// build the client object with existing API keys and API endpoints
	api_cfg := gpt3.DefaultConfig(api_key)
	api_endpoint := config.MyConfig.System.APIEndpointOpenai
	if len(api_endpoint) > 0 {
		api_cfg.BaseURL = api_endpoint
	}
	log.Println("api_cfg.BaseURL  = ", api_cfg.BaseURL)
	bot.client = gpt3.NewClientWithConfig(api_cfg)
	// bot.client.

	greeting_msg := fmt.Sprintf(config.Text("greeting_msg"), bot.name, config.Version, bot.name, strftime.Format(time.Now(), "%Y-%m-%d %H:%M:%S"))
	bot.prompt = prompt.New(
		bot.getExecutor(""),
		bot.completer,
		// prompt.OptionTitle(""),
		prompt.OptionPrefix(default_user_avatar+prompt_tip_flag),
		// prompt.OptionHistory([]string{"SELECT * FROM users;"}),
		prompt.OptionPrefixTextColor(prompt.Yellow),
		prompt.OptionPreviewSuggestionTextColor(prompt.Blue),
		prompt.OptionSelectedSuggestionBGColor(prompt.LightGray),
		prompt.OptionSuggestionBGColor(prompt.DarkGray),
	)

	// greetings once everything is ready
	bot.dumpChatHistory("\n")
	bot.Say(greeting_msg, false)

	return bot
}

func (bot *ChatBot) Run() {
	if bot.prompt != nil {
		bot.prompt.Run()
	}
}

func (bot *ChatBot) completer(doc prompt.Document) []prompt.Suggest {
	suggestions := []prompt.Suggest{}
	word := strings.ToLower(doc.TextBeforeCursor())

	suggestionsMap := get_suggestion_map(bot.role.Name)
	switch word {
	case "":
		return suggestions
	default:
		for _, s := range *suggestionsMap {
			suggestions = append(suggestions, s...)
		}
		return prompt.FilterHasPrefix(suggestions, doc.TextBeforeCursor(), true)
	}
}

func (bot *ChatBot) getExecutor(dir string) func(string) {
	return func(cmds string) {
		log.Debug("Executor running on :" + cmds)
		if len(cmds) == 0 {
			log.Debug("Blank command!")
			return
		}

		switch cmds {
		case "":
			return
		case "quit", "exit", "bye", "886", "88", "q":
			bot.Close()
			os.Exit(0)
		default:
			commandFields := strings.Fields(cmds)
			result, err := bot.CommandProcessor(cmds, commandFields)
			if err != nil {
				log.Error(err.Error())
			}
			log.Debug(result)
		}
	}
}

func (bot *ChatBot) CommandProcessor(original_msg string, arr_cmd []string) (string, error) {
	msg := ""
	if arr_cmd == nil || len(arr_cmd) == 0 {
		msg = "Invalid parameters for commandProcessor"
		log.Error(msg)
		return msg, nil
	}

	log.Debug("Executor dispatching to commander :" + arr_cmd[0])
	switch arr_cmd[0] {
	case "reset":
		log.Print(arr_cmd)

		var role string
		if len(arr_cmd) > 1 {
			role = strings.ToLower(arr_cmd[1])
		} else {
			role = config.MyConfig.System.DefaultRole
		}
		bot.resetRole(role, false)
	case "ask":
		question := original_msg[len(arr_cmd[0]):]
		if need_quit := bot.Ask(question); need_quit {
			bot.Close()
			os.Exit(0)
		}
	case "lookup":
		text := original_msg[len(arr_cmd[0]):]
		log.Debug("lookup for", text, "......")
		msg, err := utility.Lookup(text, config.MyConfig.System.PeferenceLanguage)
		if err != nil {
			if len(msg) > 0 {
				log.Error(msg)
			} else {
				log.Error(err.Error())
			}
		} else {
			log.Debug("result : " + msg)
			bot.Say(msg, false)
		}
	case "translate":
		text := original_msg[len(arr_cmd[0]):]
		log.Debug("translate for", text, "......")
		msg, err := utility.Translate(text, config.MyConfig.System.PeferenceLanguage)
		if err != nil {
			if len(msg) > 0 {
				log.Error(msg)
			} else {
				log.Error(err.Error())
			}
		} else {
			log.Debug("result : " + msg)
			bot.Say(msg, false)
		}
	case "cmd":
		if len(arr_cmd) > 1 {
			cmd_real := strings.ToLower(arr_cmd[1])
			var cmd_args []string
			if len(arr_cmd) > 2 {
				cmd_args = arr_cmd[2:]
			}
			log.Debug("cmd_real = ", cmd_real)
			log.Debug("cmd_args = ", cmd_args)

			if len(cmd_real) > 0 && cmd_real != "exit" {
				obj_cmd := exec.Command(cmd_real, cmd_args...)
				obj_cmd.Stdout = os.Stdout
				obj_cmd.Stderr = os.Stderr

				//	Run the command
				if err := obj_cmd.Run(); err != nil {
					msg = err.Error()
					bot.Say(msg, true)
					log.Error(msg)
				}
			} else {
				msg = config.Text("sys_invalid_cmd")
				bot.Say(msg, true)
				log.Error(msg)
			}
		} else {
			msg = config.Text("sys_not_enough_cmd")
			bot.Say(msg, true)
			log.Error(msg)
		}
	default:
		// log the tips for missing command
		msg := config.Text("sys_invalid_cmd") + " : " + arr_cmd[0]
		log.Info(msg)

		// treat empty commands as ask chatGPT
		if len(original_msg) > 1 {
			if need_quit := bot.Ask(original_msg); need_quit {
				bot.Close()
				os.Exit(0)
			}
		} else {
			msg = config.Text("sys_not_enough_cmd")
			bot.Say(msg, true)
			log.Error(msg)
		}
	}

	return msg, nil
}

////////////////////////////////////////////////////////////////

func (bot *ChatBot) resetRole(role_name string, keep_silent bool) {
	if role, err := config.MyConfig.FindRole(role_name); err != nil {
		bot.Say(fmt.Sprintf(config.Text("error_invalid_role"), role_name), true)
		if role, err := config.MyConfig.FindRole(config.MyConfig.System.DefaultRole); err != nil {
			if !keep_silent {
				bot.Say(fmt.Sprintf(config.Text("error_invalid_role"), role_name), true)
			}
		} else {
			bot.role = role
			if !keep_silent {
				bot.Say(fmt.Sprintf(config.Text("tips_changed_role"), config.MyConfig.System.DefaultRole, bot.role.Avatar, bot.role.Prompt), true)
			}
		}
	} else {
		bot.role = role
		if !keep_silent {
			bot.Say(fmt.Sprintf(config.Text("tips_changed_role"), role_name, bot.role.Avatar, bot.role.Prompt), true)
		}
	}

	// refresh role relevant variables
	bot.msg_history = []gpt3.ChatCompletionMessage{
		{
			Role:    "system",
			Content: bot.role.Prompt,
		},
	}
	if len(bot.role.Opening) > 0 {
		bot.msg_history = append(bot.msg_history, gpt3.ChatCompletionMessage{
			Role:    "user",
			Content: bot.role.Opening,
		})
	}
	bot.token_counter = 0

	// reopen history
	bot.initChatHistory(bot.chat_history_path, role_name)
}

func (bot *ChatBot) Close() {
	if bot.chat_history_file != nil {
		// say goodbay before closing
		bot.Say(config.Text("byebye_msg")+"\n", false)

		bot.chat_history_file.Close()
		bot.chat_history_file = nil
	}
}

func (bot *ChatBot) Say(msg string, need_dump bool) {
	utility.EchoInfo(msg)
	if need_dump {
		bot.dumpChatHistory(msg)
	}
}

func (bot *ChatBot) Ask(question string) bool {
	need_quit := false

	if len(question) <= 2 {
		msg := config.Text("sys_not_enough_cmd")
		bot.Say(msg, true)
		log.Error(msg)
		return need_quit
	}
	// add question into the conversation history
	bot.updateHistory("user", question)

	start := time.Now()
	resp, err := bot.client.CreateChatCompletion(
		context.Background(),
		gpt3.ChatCompletionRequest{
			Model:    gpt3.GPT3Dot5Turbo,
			Messages: bot.msg_history,
			// MaxTokens: 4000 - bot.token_counter,
			MaxTokens:   3000,
			Temperature: bot.role.Temperature,
			N:           1,
		},
	)
	elapsed := time.Since(start)
	log.Info("Time cost for chatGPT API request : ", elapsed)
	log.Debug(resp)

	if err != nil {
		bot.Say(err.Error(), true)
		log.Error(err)
		// need_quit = true
	} else {
		bot.token_counter = resp.Usage.TotalTokens
		message := resp.Choices[0].Message.Content

		fmt.Println(bot.role.Avatar + prompt_tip_flag)
		bot.Say(message, false)

		gray := color.New(color.FgHiBlack).PrintfFunc()
		gray(
			"%60s ( %d / %d ) %ds \n",
			strftime.Format(time.Now(), "%Y-%m-%d %H:%M:%S"),
			bot.token_counter, config.MaxTokens, elapsed/100000000,
		)
		bot.updateHistory("assistant", message)
	}

	return need_quit
}

func (bot *ChatBot) updateHistory(role string, content string) {
	// update conversation history
	bot.msg_history = append(bot.msg_history, gpt3.ChatCompletionMessage{
		Role:    role,
		Content: content,
	})

	// dump conversation history if necessary
	if role == "user" {
		bot.dumpChatHistory("#### " + content + "\n")
	} else {
		bot.dumpChatHistory(bot.role.Avatar + prompt_tip_flag + "\n")
		bot.dumpChatHistory(content + "\n")
	}
}

func (bot *ChatBot) dumpChatHistory(content string) {
	if bot.log_history && bot.chat_history_file != nil {
		size, err := bot.chat_history_file.WriteString(content)
		if err != nil {
			bot.Say(err.Error(), true)
			log.Error(err.Error())
		} else {
			// log.Debug(content)
			log.Debug("#of of bytes(", size, ") has been written to ", bot.chat_history_file.Name())
			if err = bot.chat_history_file.Sync(); err != nil {
				bot.Say(err.Error(), true)
				log.Error(err.Error())
			}
		}
	}
}

func (bot *ChatBot) initChatHistory(chat_history_path string, prefix string) bool {
	if bot.log_history {
		// create folder if necessary
		log.Debug("create folder if necessary : ", chat_history_path)
		if _, err := os.Stat(chat_history_path); os.IsNotExist(err) {
			errDir := os.MkdirAll(chat_history_path, 0755)
			if errDir != nil {
				log.Error(err)
				return false
			}
		}

		// close old one
		bot.chat_history_file.Close()
		bot.chat_history_file = nil

		// open new one
		chat_history_fname := filepath.Join(chat_history_path, prefix+"_"+utility.GetYYYYMM()+".md")
		var err error
		log.Debug("Chat History will be stored at ", chat_history_fname)
		bot.chat_history_file, err = os.OpenFile(chat_history_fname, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Println(err)
			return false
		}
	}
	return true
}
