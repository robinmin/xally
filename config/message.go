package config

var i18n_str_table = map[string]string{
	"greeting_msg": `# %s (%s)

您好，我是您的私人助理%s (%s - %s), 请问有什么可以帮您？  [ %s ]
`,
	"byebye_msg": "好的，回头见！👋🏻",

	"sys_invalid_cmd":    "无效的内部命令或者当前版本不支持",
	"sys_not_enough_cmd": "没有有效的命令，无法执行(长度大于2)",

	"error_no_chatgpt_key": "缺少环境变量 OPENAI_API_KEY, 请指定或到[openai.com](https://platform.openai.com/account/api-keys)申请。\n如需使用集中共享模式，请首先使用config-email命令注册账户。具体格式为：config-email [your_email_box] [your_service_endpoint]，[]内的内容请替换为你的实际邮箱与服务地址。",
	"error_no_deepl_key":   "缺少环境变量 DEEPL_API_KEY, 请指定或到[deepl.com](https://www.deepl.com/pro-api?cta=header-pro-api/)申请",

	"error_failed_exec":          "执行命令失败：",
	"error_invalid_role":         "指定了无效的角色(%s)，已重置为默认角色：",
	"error_invalid_email":        "指定了无效的Email(%s)，请重新输入",
	"error_invalid_endpoint_url": "指定了无效的服务器地址(%s)，请重新输入",

	"error_request_success":          "请求处理出错",
	"error_request_failed":           "请求处理成功",
	"error_invalid_params":           "参数无效",
	"error_invalid_url":              "404 : 无法路由",
	"error_invalid_http_method":      "405 : 不支持的HTTP方法",
	"error_invalid_access_denied":    "无权访问或app_token已过期",
	"error_invalid_email_register":   "注册失败, 邮箱地址无效或者该域名邮箱不被允许注册",
	"error_invalid_token":            "无效的激活码或激活码已经过期",
	"error_user_register_failed":     "注册失败",
	"error_user_register_success":    "注册成功，请前往邮箱验证账户",
	"error_user_register_success2":   "注册已完成，请联系管理员激活",
	"error_generate_token_failed":    "生成激活码/访问码失败",
	"error_send_email_failed":        "注册成功，但发送验证邮件失败。请联络管理员！",
	"tips_email_subject":             "请验证您的账户",
	"tips_email_title_activate":      "激活你的账户",
	"tips_email_ignore_msg":          "如果不是您在激活本账户，请忽略本邮件。",
	"tips_email_title_activate_ok":   "恭喜！",
	"tips_email_title_activate_ng":   "糟糕~！",
	"tips_email_content_activate":    "谢谢您的注册，请点击按钮激活您的账户",
	"tips_email_content_activate_ok": "您的账户已激活，希望您用X-Ally玩得开心.",
	"tips_email_content_activate_ng": "用户激活过程中遇到点问题，请联络管理员",
	"tips_models_shared_limited":     "目前的集中共享模式只支持gpt-3.5-turbo",
	"tips_models_failed_fetch":       "远程获取当前支持的模型失败",
	"tips_models_now_support":        "目前支持的模型包括",

	"tips_suggestion_quit":         "退出本程序",
	"tips_suggestion_reset":        "重置角色为：",
	"tips_suggestion_cmd":          "执行本地命令，并将结果回显",
	"tips_suggestion_clear":        "清除对话历史",
	"tips_suggestion_config_email": "设置Email地址",
	"tips_suggestion_ask":          "问ChatGPT",

	"tips_suggestion_file_content":      "问ChatGPT文件内容",
	"tips_suggestion_file_summary":      "文件内容摘要",
	"tips_suggestion_file_translate_cn": "文件内容翻译为中文",
	"tips_suggestion_file_translate_en": "文件内容翻译为英文",
	"tips_suggestion_file_translate_jp": "文件内容翻译为日文",

	"tips_suggestion_web_content":      "加载网页内容",
	"tips_suggestion_web_summary":      "网页内容摘要",
	"tips_suggestion_web_translate_cn": "网页内容翻译为中文",
	"tips_suggestion_web_translate_en": "网页内容翻译为英文",
	"tips_suggestion_web_translate_jp": "网页内容翻译为日文",
	"tips_suggestion_translate":        "用DeepL翻译或查字典",
	"tips_suggestion_models":           "显示当前API key支持的模型",

	"tips_changed_role":       "已为您切换为%s%s (%s), 我的提示词为：\n%s",
	"tips_not_connected":      "当前尚未链接服务端，请联系您的管理员",
	"tips_invalid_server":     "无效的服务器地址，请通过config-email命令完成设置和验证",
	"tips_no_email":           "中心化共享模式时必须有有效的Email地址，请通过config-email命令完成Email设置和验证",
	"tips_no_app_token":       "app_token无效，请通过config-email命令完成Email设置和验证。如果问题仍然持续，请联系您的管理员",
	"tips_config_email_usage": "设定邮件格式请用下面的格式: config-email [你的邮件地址] [你的服务器地址]",

	"prompt_content_summary": "请根据后文做内容摘要，并以列表的形式、尽可能精准、简明扼要地逐一列出其要点。如可能给出一句话评语",
	"prompt_translate_cn":    "请将后文内容翻译为中文，尽量做到精准地道，文中代码部分不要翻译：",
	"prompt_translate_en":    "请将后文内容翻译为英文，尽量做到精准地道，文中代码部分不要翻译：",
	"prompt_translate_jp":    "请将后文内容翻译为日文，尽量做到精准地道，文中代码部分不要翻译：",
}

var i18n_str_table_en = map[string]string{
	"greeting_msg": `# %s (%s)

Hello, I am your personal assistant %s (%s - %s), how can I help you?  [ %s ]
`,
	"byebye_msg": "Okay, see you later!👋🏻",

	"sys_invalid_cmd":    "Invalid internal command or not supported by current version",
	"sys_not_enough_cmd": "nvalid command to execute (length greater than 2)",

	"error_no_chatgpt_key": "The environment variable OPENAI_API_KEY is missing, please specify it or request it from [openai.com](https://platform.openai.com/account/api-keys). \nIf you want to use the centralized sharing mode, please use the config-email command to register an account first. The specific format is: config-email [your_email_box] [your_service_endpoint], please replace the content in [] with your actual email and service address.",
	"error_no_deepl_key":   "Missing environment variable DEEPL_API_KEY, please specify or apply at [deepl.com](https://www.deepl.com/pro-api?cta=header-pro-api/)",

	"error_failed_exec":          "Failed to execute command: ",
	"error_invalid_role":         "An invalid role (%s) was specified and has been reset to the default role: ",
	"error_invalid_email":        "An invalid Email (%s) was specified, please re-enter it",
	"error_invalid_endpoint_url": "An invalid server address (%s) was specified, please re-enter",

	"error_request_success":          "Failed to process current request",
	"error_request_failed":           "Success to process current request",
	"error_invalid_params":           "Invalid parameters",
	"error_invalid_url":              "404 : Invalid routes",
	"error_invalid_http_method":      "405 : Unsupported HTTP methods",
	"error_invalid_access_denied":    "Access denied or app_token has expired",
	"error_invalid_email_register":   "Registration failed, the email address is invalid or the domain email is not allowed to register",
	"error_invalid_token":            "Invalid activation code or activation code has expired",
	"error_user_register_failed":     "Registration failed",
	"error_user_register_success":    "Register successfully, please go to your email to verify your account",
	"error_user_register_success2":   "Registration is complete, please contact the administrator to activate",
	"error_generate_token_failed":    "Failed to generate activation token/access token",
	"error_send_email_failed":        "Registration was successful, but sending the verification email failed. Please contact your administrator!",
	"tips_email_subject":             "Please verify your account",
	"tips_email_title_activate":      "Activate Your Account",
	"tips_email_ignore_msg":          "If you did not sign up for this account, please ignore this email.",
	"tips_email_title_activate_ok":   "Congratulations!",
	"tips_email_title_activate_ng":   "Opppps!",
	"tips_email_content_activate":    "Thank you for signing up! Please click the button below to activate your account.",
	"tips_email_content_activate_ok": "Your account is ready now. Have fun with X-Ally.",
	"tips_email_content_activate_ng": "If you encounter any problems during the activation process, please contact the administrator",
	"tips_models_shared_limited":     "The current centralized sharing mode only supports gpt-3.5-turbo",
	"tips_models_failed_fetch":       "Failed to list all supported model from remote server",
	"tips_models_now_support":        "Currently supported models include",

	"tips_suggestion_quit":         "Exit",
	"tips_suggestion_reset":        "Reset role to: ",
	"tips_suggestion_cmd":          "Execute local commands and display the results",
	"tips_suggestion_clear":        "Clear all conversation hiostory",
	"tips_suggestion_config_email": "Config Email address",
	"tips_suggestion_ask":          "Ask ChatGPT",

	"tips_suggestion_file_content":      "Ask ChatGPT about file contents",
	"tips_suggestion_file_summary":      "File Content Summary",
	"tips_suggestion_file_translate_cn": "Translate file content into Chinese",
	"tips_suggestion_file_translate_en": "Translate file content into English",
	"tips_suggestion_file_translate_jp": "Translate file content into Japanese",

	"tips_suggestion_web_content":      "Load web content",
	"tips_suggestion_web_summary":      "Web Page Summary",
	"tips_suggestion_web_translate_cn": "Translate web content into Chinese",
	"tips_suggestion_web_translate_en": "Translate web content into English",
	"tips_suggestion_web_translate_jp": "Translate web content into Japanese",
	"tips_suggestion_translate":        "Use DeepL to translate or look up the dictionary",
	"tips_suggestion_models":           "Show all supported models for current API key",

	"tips_changed_role":       "Switched to %s%s (%s), my prompt : \n%s",
	"tips_not_connected":      "No connection to the server, please contact your system administrator.",
	"tips_invalid_server":     "Invalid server address, please complete the setup and verification via config-email command",
	"tips_no_email":           "A valid Email address is required for the shared mode, please complete the Email setting and verification through the config-email command.",
	"tips_no_app_token":       "The app_token is invalid, please complete the Email setup and verification via the config-email command. If the problem still persists, please contact your administrator",
	"tips_config_email_usage": "To set up email, use this command format: config-email [your email address] [your server address]",

	"prompt_content_summary": "Please make a summary of the following content and list each of its main points into bullet points as concisely as possible. If possible give a one-sentence comment: ",
	"prompt_translate_cn":    "Please translate the following content into Chinese and make it as accurate and authentic. DO NOT translate the code part of the text: ",
	"prompt_translate_en":    "Please translate the following content into English and make it as accurate and authentic. DO NOT translate the code part of the text: ",
	"prompt_translate_jp":    "Please translate the following content into Japanese and make it as accurate and authentic. DO NOT translate the code part of the text: ",
}

var i18n_str_table_jp = map[string]string{
	"greeting_msg": `# %s (%s)

こんにちは、私はあなたのパーソナルアシスタント%s (%s - %s)です、あなたのために何ができますか？  [ %s ]
`,
	"byebye_msg": "じゃあ、またね！👋🏻",

	"sys_invalid_cmd":    "不正な内部コマンドまたは現在のバージョンでサポートされていない",
	"sys_not_enough_cmd": "実行する有効なコマンドがない（長さが2以上）",

	"error_no_chatgpt_key": "環境変数OPENAI_API_KEYがありません。[openai.com](https://platform.openai.com/account/api-keys)から指定またはリクエストしてください。 \n集中共有モードを使用する場合は、まずconfig-emailコマンドでアカウントを登録してください。 形式は、config-email [your_email_box] [your_service_endpoint] で、[]内の内容は実際のメールとサービスアドレスに置き換えてください。",
	"error_no_deepl_key":   "環境変数DEEPL_API_KEYがありません。[deepl.com](https://www.deepl.com/pro-api?cta=header-pro-api/)で指定またはリクエストしてください。",

	"error_failed_exec":          "実行失敗：",
	"error_invalid_role":         "無効なロール (%s) が指定されたので、デフォルトのロールにリセットされました：",
	"error_invalid_email":        "無効な電子メール(%s)が指定されました、再度入力してください。",
	"error_invalid_endpoint_url": "無効なサーバーアドレス(%s)が指定されました、再度入力してください。",

	"error_request_success":          "現在の要求の処理に失敗しました",
	"error_request_failed":           "現在のリクエストを処理することに成功",
	"error_invalid_params":           "無効なパラメータ",
	"error_invalid_url":              "404 : 無効なルート",
	"error_invalid_http_method":      "405 : サポートされていないHTTPメソッド",
	"error_invalid_access_denied":    "アクセス不可か、app_tokenの有効期限が切れています。",
	"error_invalid_email_register":   "登録に失敗しました。メールアドレスが無効か、ドメインメールの登録が許可されていません。",
	"error_invalid_token":            "無効なアクティベーションコード、またはアクティベーションコードの有効期限が切れています。",
	"error_user_register_failed":     "登録に失敗しました",
	"error_user_register_success":    "登録に成功しました。アカウントの確認のため、メールにアクセスしてください。",
	"error_user_register_success2":   "登録が完了しましたので、管理者に連絡してアクティベーションを行ってください",
	"error_generate_token_failed":    "アクティベーショントークン/アクセストークンの生成に失敗しました。",
	"error_send_email_failed":        "登録は成功しましたが、認証メールの送信に失敗しました。 管理者までご連絡ください！",
	"tips_email_subject":             "アカウントの確認をしてください",
	"tips_email_title_activate":      "アカウントの有効化",
	"tips_email_ignore_msg":          "このアカウントに登録されていない方は、このメールを無視してください。",
	"tips_email_title_activate_ok":   "おめでとうございます！",
	"tips_email_title_activate_ng":   "オッパッピー！",
	"tips_email_content_activate":    "ご登録いただきありがとうございます！下のボタンをクリックして、アカウントを有効にしてください。",
	"tips_email_content_activate_ok": "あなたのアカウントは今準備が整っています。X-Allyで楽しんでください。",
	"tips_email_content_activate_ng": "アクティベーション中に問題が発生した場合は、管理者までご連絡ください",
	"tips_models_shared_limited":     "現在の集中共有モデルは、gpt-3.5-turboにのみ対応しています",
	"tips_models_failed_fetch":       "リモートサーバーの対応モデル一覧に失敗しました",
	"tips_models_now_support":        "現在対応しているモデルは以下の通り",

	"tips_suggestion_quit":         "終了する",
	"tips_suggestion_reset":        "役割をリセットして：",
	"tips_suggestion_cmd":          "ローカルコマンドを実行し、その結果を表示する",
	"tips_suggestion_clear":        "対話履歴をクリアにする",
	"tips_suggestion_config_email": "コンフィグ電子メール",
	"tips_suggestion_ask":          "ChatGPTに問い合わせて",

	"tips_suggestion_file_content":      "文書内容をChatGPTに聞く",
	"tips_suggestion_file_summary":      "文書概要を纏める",
	"tips_suggestion_file_translate_cn": "文書内容を中国語への翻訳",
	"tips_suggestion_file_translate_en": "文書内容を英語への翻訳",
	"tips_suggestion_file_translate_jp": "文書内容を日本語への翻訳",

	"tips_suggestion_web_content":      "ウェブページ内容を読み込む",
	"tips_suggestion_web_summary":      "ウェブページ内容概要を纏める",
	"tips_suggestion_web_translate_cn": "ウェブページ内容を中国語への翻訳",
	"tips_suggestion_web_translate_en": "ウェブページ内容を英語への翻訳",
	"tips_suggestion_web_translate_jp": "ウェブページ内容を日本語への翻訳",
	"tips_suggestion_translate":        "DeepLで翻訳する、または辞書を調べて",
	"tips_suggestion_models":           "APIキーが現在サポートしているモデルを表示する",

	"tips_changed_role":       "%s%s (%s)に切り替えました、私のプロンプトワードは : \n%s",
	"tips_not_connected":      "サーバーと接続していないのため、システム管理者にお問い合わせください",
	"tips_invalid_server":     "サーバーアドレスが無効です。config-emailコマンドで設定と確認を完了してください。",
	"tips_no_email":           "集中共有モードでは、有効な電子メールアドレスが必要です。config-emailコマンドを使用して、電子メールアドレスの設定と確認をしてください。",
	"tips_no_app_token":       "app_tokenが無効です。config-emailコマンドでEmailの設定と検証を完了してください。 問題が解決しない場合は、管理者に連絡してください。",
	"tips_config_email_usage": "メールの設定は、次のコマンド形式で行います：config-email [あなたのメールアドレス] [あなたのサーバーアドレス]。",

	"prompt_content_summary": "以下の内容を要約し、それぞれの要点をできるだけ簡潔に箇条書きにしてください。可能であれば、1文のコメントを添えてください：",
	"prompt_translate_cn":    "後者をできるだけ正確に中国語に翻訳し、コード部分は翻訳しないようにしてください：",
	"prompt_translate_en":    "後者をできるだけ正確に英語に翻訳し、コード部分は翻訳しないようにしてください：",
	"prompt_translate_jp":    "後者をできるだけ正確に日本語に翻訳し、コード部分は翻訳しないようにしてください：",
}
