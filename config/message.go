package config

var i18n_str_table = map[string]string{
	// "greeting_msg": "您好，我是您的私人助理x-ally, 请问有什么可以帮您？帮助请输入@help、退出请输入@quit",
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

	"tips_suggestion_quit":      "退出本程序",
	"tips_suggestion_reset":     "重置机器人角色为：",
	"tips_suggestion_cmd":       "执行本地命令，并将结果回显",
	"tips_suggestion_ask":       "问chatGPT",
	"tips_suggestion_translate": "用deepl翻译或查字典",
	"tips_changed_role":         "已为您切换为%s%s, %s",
}

// //////////////////////////////////////////////////////////////////////////////
func Text(str_key string) string {
	if str_val, ok := i18n_str_table[str_key]; ok {
		return str_val
	} else {
		return str_key
	}
}
