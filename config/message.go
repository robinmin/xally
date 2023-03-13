package config

var i18n_str_table = map[string]string{
	"greeting_msg": `# %s (%s)

您好，我是您的私人助理%s, 请问有什么可以帮您？  [ %s ]
`,
	"byebye_msg": "好的，回头见！👋🏻",

	"sys_invalid_cmd":    "无效的内部命令或者当前版本不支持",
	"sys_not_enough_cmd": "没有有效的命令，无法执行(长度大于2)",

	"error_no_chatgpt_key": "缺少环境变量 OPENAI_API_KEY, 请指定或到[openai.com](https://platform.openai.com/account/api-keys)申请",
	"error_no_deepl_key":   "缺少环境变量 DEEPL_API_KEY, 请指定或到[deepl.com](https://www.deepl.com/pro-api?cta=header-pro-api/)申请",

	"error_failed_exec":  "执行命令失败：",
	"error_invalid_role": "指定了无效的角色(%s)，已重置为默认角色：",

	"tips_suggestion_quit":        "退出本程序",
	"tips_suggestion_reset":       "重置角色为：",
	"tips_suggestion_cmd":         "执行本地命令，并将结果回显",
	"tips_suggestion_ask":         "问ChatGPT",
	"tips_suggestion_file":        "问ChatGPT文件内容",
	"tips_suggestion_web_content": "加载网页内容",
	"tips_suggestion_web_summary": "网页内容摘要",
	"tips_suggestion_translate":   "用DeepL翻译或查字典",
	"tips_changed_role":           "已为您切换为%s%s, 我的提示词为：\n%s",
}

var i18n_str_table_en = map[string]string{
	"greeting_msg": `# %s (%s)

Hello, I am your personal assistant %s, how can I help you?  [ %s ]
`,
	"byebye_msg": "Okay, see you later!👋🏻",

	"sys_invalid_cmd":    "Invalid internal command or not supported by current version",
	"sys_not_enough_cmd": "nvalid command to execute (length greater than 2)",

	"error_no_chatgpt_key": "Missing environment variable OPENAI_API_KEY, please specify or apply at [openai.com](https://platform.openai.com/account/api-keys)",
	"error_no_deepl_key":   "Missing environment variable DEEPL_API_KEY, please specify or apply at [deepl.com](https://www.deepl.com/pro-api?cta=header-pro-api/)",

	"error_failed_exec":  "Failed to execute command: ",
	"error_invalid_role": "An invalid role (%s) was specified and has been reset to the default role: ",

	"tips_suggestion_quit":        "Exit",
	"tips_suggestion_reset":       "Reset role to: ",
	"tips_suggestion_cmd":         "Execute local commands and display the results",
	"tips_suggestion_ask":         "Ask ChatGPT",
	"tips_suggestion_file":        "Ask ChatGPT about file contents",
	"tips_suggestion_web_content": "Load web content",
	"tips_suggestion_web_summary": "Web Page Summary",
	"tips_suggestion_translate":   "Use DeepL to translate or look up the dictionary",
	"tips_changed_role":           "Switched to %s%s, my prompt : \n%s",
}

var i18n_str_table_jp = map[string]string{
	"greeting_msg": `# %s (%s)

こんにちは、私はあなたのパーソナルアシスタント%sです、あなたのために何ができますか？  [ %s ]
`,
	"byebye_msg": "じゃあ、またね！👋🏻",

	"sys_invalid_cmd":    "不正な内部コマンドまたは現在のバージョンでサポートされていない",
	"sys_not_enough_cmd": "実行する有効なコマンドがない（長さが2以上）",

	"error_no_chatgpt_key": "環境変数OPENAI_API_KEYが不足しています。指定するか、[openai.com](https://platform.openai.com/account/api-keys)からリクエストしてください。",
	"error_no_deepl_key":   "環境変数DEEPL_API_KEYがありません。[deepl.com](https://www.deepl.com/pro-api?cta=header-pro-api/)で指定またはリクエストしてください。",

	"error_failed_exec":  "実行失敗：",
	"error_invalid_role": "無効なロール (%s) が指定されたので、デフォルトのロールにリセットされました：",

	"tips_suggestion_quit":        "終了する",
	"tips_suggestion_reset":       "役割をリセットして：",
	"tips_suggestion_cmd":         "ローカルコマンドを実行し、その結果を表示する",
	"tips_suggestion_ask":         "ChatGPTに問い合わせて",
	"tips_suggestion_file":        "ChatGPTに資料の内容を聞く",
	"tips_suggestion_web_content": "Webコンテンツの読み込み",
	"tips_suggestion_web_summary": "ページ内容の概要",
	"tips_suggestion_translate":   "DeepLで翻訳する、または辞書を調べて",
	"tips_changed_role":           "%s%sに切り替えました、私のプロンプトワードは : \n%s",
}
