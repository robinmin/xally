package cmd

import (
	"context"
	"encoding/json"
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

	"xhqb.com/tools/xally/config"
)

const default_user_avatar = "👦"
const default_ai_role = "expert"

type suggestionType int

const (
	// execute host command on locase machine
	Ask suggestionType = iota

	// execute host command on locase machine
	Cmd

	// Reset suggestion
	Reset

	// Quit suggestion
	Quit

	// Others is key for various arbitrary suggestions
	Others
)

var suggestionsMap = map[suggestionType][]prompt.Suggest{
	Ask: {
		{Text: "ask", Description: config.Text("tips_suggestion_ask")},
	},
	Cmd: {
		{Text: "cmd", Description: config.Text("tips_suggestion_cmd")},
	},
	Reset: {
		{Text: "reset expert", Description: config.Text("tips_suggestion_reset") + "expert"},
		{Text: "reset assistant", Description: config.Text("tips_suggestion_reset") + "assistant"},
	},
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

type SysRole struct {
	role        string
	avatar      string
	description string
	temperature float32
}

var predefined_roles = map[string]SysRole{
	"expert": {
		role:        "system",
		avatar:      "🐬",
		description: "You are ChatGPT, a large language model trained by OpenAI. Answer as concisely as possible. Knowledge cutoff: {knowledge_cutoff} Current date: {current_date}",
		temperature: 0.2,
	},
	"assistant": {
		role:        "system",
		avatar:      "🧰",
		description: "You are a ChatGPT-based daily chit-chat bot with answers that are as concise and soft as possible.",
		temperature: 1.8,
	},
}

type ChatBot struct {
	// chatbot name
	name string
	role SysRole

	// logfile for chat conversation
	chat_history_file *os.File
	chat_history_path string

	client        *gpt3.Client
	msg_history   []gpt3.ChatCompletionMessage
	log_history   bool
	token_counter int
	prompt        *prompt.Prompt
}

func NewChatbot(chat_history_path string, name string, log_history bool) *ChatBot {
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
	bot.resetRole(default_ai_role)

	log.Debug("bot.log_history = ", bot.log_history)
	if bot.log_history {
		// create folder if necessary
		log.Debug("create folder if necessary : ", chat_history_path)
		if _, err := os.Stat(chat_history_path); os.IsNotExist(err) {
			errDir := os.MkdirAll(chat_history_path, 0755)
			if errDir != nil {
				log.Error(err)
				return bot
			}
		}

		chat_history_fname := filepath.Join(chat_history_path, bot.name+"_"+get_yyyymmdd()+".md")
		var err error
		log.Debug("Chat History will be stored at ", chat_history_fname)
		bot.chat_history_file, err = os.OpenFile(
			chat_history_fname,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY,
			0644,
		)
		if err != nil {
			log.Println(err)
		}
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		bot.Say("- "+config.Text("error_no_chatgpt_key"), true)
		return bot
	}

	// scanner := bufio.NewScanner(os.Stdin)
	bot.client = gpt3.NewClient(apiKey)

	greeting_msg := fmt.Sprintf(config.Text("greeting_msg"), bot.name, config.Version, bot.name, strftime.Format(time.Now(), "%Y-%m-%d %H:%M:%S"))
	bot.prompt = prompt.New(
		bot.getExecutor(""),
		bot.completer,
		// prompt.OptionTitle(""),
		prompt.OptionPrefix(default_user_avatar+" ▶ "),
		// prompt.OptionHistory([]string{"SELECT * FROM users;"}),
		prompt.OptionPrefixTextColor(prompt.Yellow),
		prompt.OptionPreviewSuggestionTextColor(prompt.Blue),
		prompt.OptionSelectedSuggestionBGColor(prompt.LightGray),
		prompt.OptionSuggestionBGColor(prompt.DarkGray),
	)

	// greetings once everything is ready
	bot.dumpConversation("\n")
	bot.Say(greeting_msg, true)

	return bot
}

func (bot *ChatBot) Run() {
	if bot.prompt != nil {
		bot.prompt.Run()
	}
}

func (bot *ChatBot) completer(d prompt.Document) []prompt.Suggest {
	suggestions := []prompt.Suggest{}
	word := strings.ToLower(d.TextBeforeCursor())

	switch word {
	case "":
		return suggestions
	// case "reset":
	// 	for role_name := range predefined_roles {
	// 		suggestions = append(suggestions, prompt.Suggest{
	// 			Text:        "reset " + role_name,
	// 			Description: config.Text("tips_suggestion_reset") + role_name,
	// 		})
	// 	}
	// 	return suggestions
	default:
		for _, s := range suggestionsMap {
			suggestions = append(suggestions, s...)
		}
		return prompt.FilterHasPrefix(suggestions, d.TextBeforeCursor(), true)
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
			result, err := bot.commandProcessor(cmds, commandFields)
			if err != nil {
				log.Error(err.Error())
			}
			log.Debug(result)
		}
	}
}

func (bot *ChatBot) commandProcessor(original_msg string, arr_cmd []string) (string, error) {
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

		role := default_ai_role
		if len(arr_cmd) > 1 {
			role = strings.ToLower(arr_cmd[1])
		}
		bot.resetRole(role)
	case "ask":
		question := original_msg[len(arr_cmd[0]):]
		if need_quit := bot.Ask(question); need_quit {
			bot.Close()
			os.Exit(0)
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

func (bot *ChatBot) resetRole(role_name string) {
	// find the role name
	role, ok := predefined_roles[role_name]
	if ok {
		bot.role = role
		bot.Say(fmt.Sprintf(config.Text("tips_changed_role"), role_name, bot.role.avatar, bot.role.description), true)
	} else {
		bot.role = predefined_roles[default_ai_role]
		bot.Say(fmt.Sprintf(config.Text("error_invalid_role"), role_name), true)
	}

	// refresh role relevant variables
	bot.msg_history = []gpt3.ChatCompletionMessage{
		{
			Role:    bot.role.role,
			Content: bot.role.description,
		},
	}
	bot.token_counter = 0
}

func (bot *ChatBot) Close() {
	if bot.chat_history_file != nil {
		// say goodbay before closing
		bot.Say(config.Text("byebye_msg")+"\n", true)

		bot.chat_history_file.Close()
		bot.chat_history_file = nil
	}
}

func (bot *ChatBot) Say(msg string, need_dump bool) {
	echo_info(msg)
	if need_dump {
		bot.dumpConversation(msg)
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
			Temperature: bot.role.temperature,
			N:           1,
		},
	)
	elapsed := time.Since(start)
	log.Info("Time cost for chatGPT API request : ", elapsed)
	// log.Debug(resp)
	json_resp, err1 := json.Marshal(resp)
	if err1 != nil {
		fmt.Printf("Error: %s", err1)
	}
	log.Debug(json_resp)

	if err != nil {
		bot.Say(err.Error(), true)
		log.Error(err)
		// need_quit = true
	} else {
		bot.token_counter = resp.Usage.TotalTokens
		message := resp.Choices[0].Message.Content

		fmt.Println(bot.role.avatar + " ▶ ")
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
	converasation_mark := ""
	if role == "user" {
		converasation_mark = "#### " + default_user_avatar
	} else {
		converasation_mark = bot.role.avatar
	}
	new_content := "\n" + converasation_mark + " : " + content
	bot.dumpConversation(new_content)
}

func (bot *ChatBot) dumpConversation(content string) {
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
