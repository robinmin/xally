package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/c-bata/go-prompt"
	"github.com/fatih/color"
	strftime "github.com/itchyny/timefmt-go"
	gpt3 "github.com/sashabaranov/go-openai"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"

	"github.com/robinmin/xally/config"
	"github.com/robinmin/xally/shared/clientdb"
	"github.com/robinmin/xally/shared/model"
	"github.com/robinmin/xally/shared/utility"
)

const default_user_avatar = "🧑"
const prompt_tip_flag = " ▶ "

type suggestionType int

const (
	// execute host command on local machine
	Ask suggestionType = iota

	// translate text
	Translate

	// data from web page contents
	Web

	// data from file contents
	File

	// execute host command on local machine
	Cmd

	// Reset suggestion
	Reset

	// Quit suggestion
	Quit

	// Others is key for various arbitrary suggestions
	Others
)

func get_suggestion_map(role_name string) *map[suggestionType][]prompt.Suggest {
	// culmulate role list
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
		Web: {
			{Text: PLUGIN_NAME_WEB_CONTENT, Description: config.Text("tips_suggestion_web_content")},
			{Text: PLUGIN_NAME_WEB_SUMMARY, Description: config.Text("tips_suggestion_web_summary")},
			{Text: PLUGIN_NAME_WEB_TRANSLATE_CN, Description: config.Text("tips_suggestion_web_translate_cn")},
			{Text: PLUGIN_NAME_WEB_TRANSLATE_EN, Description: config.Text("tips_suggestion_web_translate_en")},
			{Text: PLUGIN_NAME_WEB_TRANSLATE_JP, Description: config.Text("tips_suggestion_web_translate_jp")},
		},
		File: {
			{Text: PLUGIN_NAME_FILE_CONTENT, Description: config.Text("tips_suggestion_file_content")},
			{Text: PLUGIN_NAME_FILE_SUMMARY, Description: config.Text("tips_suggestion_file_summary")},
			{Text: PLUGIN_NAME_FILE_TRANSLATE_CN, Description: config.Text("tips_suggestion_file_translate_cn")},
			{Text: PLUGIN_NAME_FILE_TRANSLATE_EN, Description: config.Text("tips_suggestion_file_translate_en")},
			{Text: PLUGIN_NAME_FILE_TRANSLATE_JP, Description: config.Text("tips_suggestion_file_translate_jp")},
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

type LivePrefixState struct {
	LivePrefix string
	IsEnable   bool
	Buffer     string
}

func (lps *LivePrefixState) ChangeLivePrefix() (string, bool) {
	return default_user_avatar + prompt_tip_flag, lps.IsEnable
}

func (lps *LivePrefixState) InputModePadding(cmds string) bool {
	// add keyboard padding to support multiple line input
	if strings.HasSuffix(cmds, ";") || strings.HasSuffix(cmds, "；") {
		lps.Buffer = lps.Buffer + cmds
		lps.IsEnable = false
		// lps.LivePrefix = cmds
		// fmt.Printf("Query: %s\n", query)
		lps.Buffer = ""
		return lps.IsEnable
	}
	lps.Buffer = lps.Buffer + cmds + " "
	// lps.LivePrefix = "..."
	lps.IsEnable = true

	return lps.IsEnable
}

func (lps *LivePrefixState) ResetInputMode() {
	lps.Buffer = ""
	lps.IsEnable = true
}

type ChatBot struct {
	// chatbot name
	name string
	role *config.SysRole

	// logfile for chat conversation
	chat_history_file *os.File
	chat_history_path string

	client                   *ChatGPTCLient
	msg_history              []gpt3.ChatCompletionMessage
	log_history              bool
	token_counter_total      int
	token_counter_completion int
	token_counter_prompt     int
	prompt                   *prompt.Prompt
	clientdb                 *clientdb.ClientDB

	plugin_mgr *PluginManager
	kb_padding *LivePrefixState
}

func NewChatbot(chat_history_path string, name string, role_name string, log_history bool, verbose bool) *ChatBot {
	bot := &ChatBot{
		name:                     name,
		chat_history_file:        nil,
		chat_history_path:        chat_history_path,
		client:                   nil,
		msg_history:              nil,
		log_history:              log_history,
		token_counter_total:      0,
		token_counter_completion: 0,
		token_counter_prompt:     0,
		prompt:                   nil,
	}
	if role_name == "" {
		bot.resetRole(config.MyConfig.System.DefaultRole, true)
	} else {
		bot.resetRole(role_name, true)
	}

	// initialize all plugins and plugin manager
	bot.plugin_mgr = NewPluginManager()
	bot.plugin_mgr.Open()

	bot.clientdb, _ = clientdb.InitClientDB(path.Join(config.MyConfig.System.ChatHistoryPath, "xally.db"), verbose)

	api_key := config.MyConfig.System.OpenaiApiKey
	if !config.MyConfig.IsSharedMode() && api_key == "" {
		bot.Say("- "+config.Text("error_no_chatgpt_key"), true)
		return bot
	}

	// build the client object with existing API keys and API endpoints
	api_cfg := gpt3.DefaultConfig(api_key)
	api_endpoint := config.MyConfig.System.APIEndpointOpenai
	if len(api_endpoint) > 0 {
		api_cfg.BaseURL = api_endpoint
	}
	log.Println("api_cfg.BaseURL  = ", api_cfg.BaseURL)
	// bot.client = gpt3.NewClientWithConfig(api_cfg)
	bot.client = &ChatGPTCLient{
		Client:     *gpt3.NewClientWithConfig(api_cfg),
		HTTPClient: api_cfg.HTTPClient,
	}

	flags := strftime.Format(time.Now(), "%m-%d %H:%M ") + config.MyConfig.GetCurrentMode(bot.CheckConnectivity())
	greeting_msg := fmt.Sprintf(config.Text("greeting_msg"), bot.name, config.Version, bot.name, flags)

	var option_history []string
	var err error
	if bot.clientdb != nil {
		if option_history, err = bot.clientdb.LoadOptionHistory(bot.role.Name); err != nil {
			log.Error("Failed to load option history: ", err)
		}
	}

	// add keyboard padding to support multiple lines when inputting
	bot.kb_padding = &LivePrefixState{}
	bot.kb_padding.ResetInputMode()

	bot.prompt = prompt.New(
		bot.getExecutor(""),
		bot.completer,
		prompt.OptionTitle(bot.name+" / "+config.Version),
		// prompt.OptionPrefix(default_user_avatar+prompt_tip_flag),
		prompt.OptionPrefix("... "),
		prompt.OptionPrefixTextColor(prompt.Yellow),
		prompt.OptionLivePrefix(bot.kb_padding.ChangeLivePrefix),
		prompt.OptionHistory(option_history),
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
		if cmds == "" {
			log.Debug("Blank command!")
			return
		}

		// add keyboard padding to support multiple line input
		if !bot.kb_padding.InputModePadding(cmds) {
			log.Debug("Enter into multiple line mo")
			return
		}

		switch cmds {
		case "":
			return
		case "quit", "exit", "bye", "886", "88", "q":
			bot.Close(true)
		default:
			commandFields := strings.Fields(cmds)
			msg, need_dump, err := bot.CommandProcessor(cmds, commandFields)
			if err != nil {
				log.Error(err.Error())
			} else {
				if len(msg) > 0 {
					bot.Say(msg, need_dump)
				}
			}
		}
	}
}

func (bot *ChatBot) CommandProcessor(original_msg string, arr_cmd []string) (string, bool, error) {
	msg := ""
	need_dump := false
	var err error

	if original_msg == "" {
		return msg, need_dump, errors.New("Invalid parameters for commandProcessor")
	}

	log.Debug("Executor dispatching to commander :" + original_msg)
	if bot.clientdb != nil {
		bot.clientdb.AddOptionHistory(&clientdb.OptionHistory{
			Role:   bot.role.Name,
			Option: original_msg,
		})
	}

	// switch to plugin manager to translate the command and content
	tmp_processed, tmp_replaced_msg, tmp_replaced_cmd, tmp_err := bot.plugin_mgr.Execute(original_msg, arr_cmd)
	if tmp_processed {
		if tmp_err == nil {
			original_msg = tmp_replaced_msg
			arr_cmd = tmp_replaced_cmd
		} else {
			msg = "[ERROR]" + tmp_err.Error()
			bot.Say("[ERROR]"+tmp_err.Error(), true)
			return msg, need_dump, tmp_err
		}
	}

	if arr_cmd == nil || len(arr_cmd) <= 0 {
		return msg, need_dump, errors.New("Invalid updated parameters for commandProcessor")
	}

	switch arr_cmd[0] {
	case "reset":
		log.Debug("Execute [reset] command on : ", original_msg)

		var role string
		if len(arr_cmd) > 1 {
			role = strings.ToLower(arr_cmd[1])
		} else {
			role = config.MyConfig.System.DefaultRole
		}
		bot.resetRole(role, false)
	case "ask":
		if len(original_msg) > len(arr_cmd[0]) {
			log.Debug("Execute [ask] command on : ", original_msg)

			question := original_msg[len(arr_cmd[0]):]
			if need_quit := bot.Ask(question); need_quit {
				bot.Close(true)
			}
		}

	case PLUGIN_NAME_FILE_CONTENT:
		if len(original_msg) > len(arr_cmd[0]) {
			log.Debug("Execute [file-content] command on : ", original_msg)

			bot.Say("> "+strings.ReplaceAll(original_msg, "\n", "\n> ")+"\n", true)
		}
	case PLUGIN_NAME_FILE_SUMMARY:
		fallthrough
	case PLUGIN_NAME_FILE_TRANSLATE_CN:
		fallthrough
	case PLUGIN_NAME_FILE_TRANSLATE_EN:
		fallthrough
	case PLUGIN_NAME_FILE_TRANSLATE_JP:
		if len(original_msg) > len(arr_cmd[0]) {
			log.Debug("Execute [%s] command on : ", arr_cmd[0], original_msg)

			bot.Say("> "+strings.ReplaceAll(original_msg, "\n", "\n> ")+"\n", true)
			if need_quit := bot.Ask(original_msg); need_quit {
				bot.Close(true)
			}
		}

	case PLUGIN_NAME_WEB_CONTENT:
		if len(original_msg) > len(arr_cmd[0]) {
			log.Debug("Execute [web-content] command on : ", original_msg)

			bot.Say("> "+strings.ReplaceAll(original_msg, "\n", "\n> ")+"\n", true)
		}
	case PLUGIN_NAME_WEB_SUMMARY:
		fallthrough
	case PLUGIN_NAME_WEB_TRANSLATE_CN:
		fallthrough
	case PLUGIN_NAME_WEB_TRANSLATE_EN:
		fallthrough
	case PLUGIN_NAME_WEB_TRANSLATE_JP:
		if len(original_msg) > len(arr_cmd[0]) {
			log.Debug("Execute [%] command on : ", arr_cmd[0], original_msg)

			bot.Say("> "+strings.ReplaceAll(original_msg, "\n", "\n> ")+"\n", true)
			if need_quit := bot.Ask(original_msg); need_quit {
				bot.Close(true)
			}
		}

	case "lookup":
		log.Debug("Execute [lookup] command on : ", original_msg)

		question := original_msg[len(arr_cmd[0]):]
		msg, err = utility.Lookup(question, config.MyConfig.System.PeferenceLanguage)
		if err == nil {
			need_dump = true
		}
	case "translate":
		log.Debug("Execute [translate] command on : ", original_msg)

		question := original_msg[len(arr_cmd[0]):]
		msg, err = utility.Translate(question, config.MyConfig.System.PeferenceLanguage)
		if err == nil {
			need_dump = true
		}
	case "cmd":
		log.Debug("Execute [cmd] command on : ", original_msg)

		if len(arr_cmd) > 1 {
			var cmd_args []string

			cmd_real := strings.ToLower(arr_cmd[1])
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
				if err = obj_cmd.Run(); err != nil {
					msg = err.Error()
					need_dump = true
				}
			} else {
				msg = config.Text("sys_invalid_cmd")
				need_dump = true
				err = errors.New(msg)
			}
		} else {
			msg = ""
			need_dump = true
			err = errors.New(config.Text("sys_not_enough_cmd"))
		}
	default:
		log.Debug("Execute fallback command on : ", original_msg)

		// treat empty commands as ask chatGPT
		if len(original_msg) > 1 {
			if need_quit := bot.Ask(original_msg); need_quit {
				bot.Close(true)
			}
		} else {
			msg = config.Text("sys_not_enough_cmd")
			need_dump = true
			err = errors.New(msg)
		}
	}

	return msg, need_dump, err
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
	bot.token_counter_total = 0
	bot.token_counter_completion = 0
	bot.token_counter_prompt = 0

	// reopen history
	bot.initChatHistory(bot.chat_history_path, role_name)
}

func (bot *ChatBot) Close(exit bool) {
	bot.plugin_mgr.Close()

	if bot.chat_history_file != nil {
		// say goodbay before closing
		bot.Say(config.Text("byebye_msg")+"\n", false)

		bot.chat_history_file.Close()
		bot.chat_history_file = nil
	}

	if exit {
		os.Exit(0)
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

	// avoid token length beyond openai limitation
	var token_len int
	var init_msg_len int
	question_len := len(question)

	if len(bot.role.Opening) > 0 {
		init_msg_len = 2
	} else {
		init_msg_len = 1
	}

	// adjust available max token length
	for {
		token_len = bot.estimateAvailableTokenNumber(question_len)
		if token_len > 0 {
			break
		}

		if len(bot.msg_history) < init_msg_len {
			fmt.Println("🌸🌸🌸🌸🌸🌸🌸🌸🌸🌸🌸🌸🌸🌸🌸🌸 resetting role......")
			// we assume the default configuration is fine
			bot.resetRole(bot.role.Name, true)
			token_len = bot.estimateAvailableTokenNumber(len(question))
			break
		} else {
			fmt.Print("🌻🌻🌻🌻🌻🌻🌻🌻🌻🌻🌻🌻🌻🌻🌻🌻 erasing old history......")
			// popup the old conversation history but key the prompt and the potential opening
			size_before := len(bot.msg_history)
			bot.msg_history = slices.Delete(bot.msg_history, init_msg_len, init_msg_len+1)
			fmt.Printf("%d --> %d\n", size_before, len(bot.msg_history))
		}
	}

	utility.ReportEvent(utility.EVT_CLIENT_ASK_CHATGPT, "Asking to chatGPT", nil)

	rqst := gpt3.ChatCompletionRequest{
		Model:     gpt3.GPT3Dot5Turbo,
		Messages:  bot.msg_history,
		MaxTokens: token_len,
		// MaxTokens:   3000,
		Temperature: bot.role.Temperature,
		N:           1,
	}
	resp, err := bot.client.CreateChatCompletion(context.Background(), rqst)

	// add chat history
	var username string
	if current_user, err := user.Current(); err == nil {
		username = current_user.Username
	} else {
		username = ""
	}
	chat_history := model.ConversationHistory{}
	chat_history.LoadRequest(bot.role.Name, username, &rqst)
	chat_history.LoadResponse(&resp)
	if !bot.clientdb.AddChatHistory(&chat_history) {
		log.Error("Failed to write chat history into local database.")
	}

	elapsed := time.Since(start)
	log.Info("Time cost for chatGPT API request : ", elapsed)
	// log.Debug(resp)
	utility.ReportEvent(utility.EVT_CLIENT_ANSWER_CHATGPT, "Answered from chatGPT", nil)

	if err != nil {
		bot.Say(err.Error(), true)
		log.Error(err)
		// need_quit = true
	} else {
		bot.token_counter_total = resp.Usage.TotalTokens
		bot.token_counter_completion = resp.Usage.CompletionTokens
		bot.token_counter_prompt = resp.Usage.PromptTokens

		message := resp.Choices[0].Message.Content

		fmt.Println(bot.role.Avatar + prompt_tip_flag)
		bot.Say(message, false)

		gray := color.New(color.FgHiBlack).PrintfFunc()
		gray(
			"%40s ( %d + %d = %d ) %ds       %s : %d\n\n",
			strftime.Format(time.Now(), "%Y-%m-%d %H:%M:%S"),
			bot.token_counter_prompt,
			bot.token_counter_completion,
			bot.token_counter_total,
			elapsed/100000000,
			strings.Repeat("░", len(bot.msg_history)),
			bot.estimateAvailableTokenNumber(question_len),
		)
		bot.updateHistory("assistant", message)

		bot.kb_padding.ResetInputMode()
	}

	return need_quit
}

func (bot *ChatBot) estimateAvailableTokenNumber(question_leng int) int {
	available_len := config.MaxTokens - question_leng // - bot.token_counter_total

	for _, msg := range bot.msg_history {
		available_len = available_len - 4 // every message follows <im_start>{role/name}\n{content}<im_end>\n
		// TODO: need to fine tune going forward, replaced with GPT-index instead of length of contents
		available_len = available_len - len(msg.Content) - len(msg.Name)
		available_len = available_len - 1 // - len(msg.Role), role is always required and always 1 token
		available_len = available_len - 2 // every reply is primed with <im_start>assistant
	}

	return available_len
}

func (bot *ChatBot) updateHistory(role string, content string) {
	// update conversation history
	bot.msg_history = append(bot.msg_history, gpt3.ChatCompletionMessage{
		Role:    role,
		Content: content,
	})

	// dump conversation history if necessary
	if role == "user" {
		bot.dumpChatHistory("#### " + default_user_avatar + "  " + content + "\n")
	} else {
		bot.dumpChatHistory(bot.role.Avatar + prompt_tip_flag + "\n")
		bot.dumpChatHistory(content + "\n\n")
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

func (bot *ChatBot) CheckConnectivity() bool {
	// always return true when non-shared mode
	if config.MyConfig.System.UseSharedMode == 0 {
		return true
	}

	// check app_token is valid or not
	if len(config.MyConfig.System.AppToken) > 0 {
		// TODO : call remote RPC to check app_token is valid or not
		return false
	}

	// invalid app_token when shared mode
	return false
}
