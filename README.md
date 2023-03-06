# X-Ally

(v0.0.4)

X-Ally is an AI-based TUI (aka Terminal User Interface) tool that helps people do things more elegantly. So far it has been integrated with the APIs from [openai.com](https://openai.com/) and [deepl.com](https://www.deepl.com/).


#### Installation
You can directly download the latest version of X-Ally from [here](https://github.com/robinmin/xally/releases/). If you also use macOS, you can install the latest version X-Ally via brew as shown below:

```bash
brew tap robinmin/homebrew-tap
brew install xally
```

#### Usage
You need to setup your environment variables before using. For example(Please replace your real keys with the following dummys):

```bash
# key from openai.com, mandatory so far
export OPENAI_API_KEY=sk-XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
# key from deepl.com, optional
export DEEPL_API_KEY=XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX

# run the program
$ xally --help
```

then you will get the following tips:
```bash
xally version: xally/x.x.xx
Usage: xally [-hl] [-p history_path] [-w language_preference]

Options:
  -h	show the help message
  -l	flag to log history (default true)
  -p string
    	specify chat history path (default "data")
  -w string
    	language preference, so far only support CN, JP and EN (default "CN")
```


![xally_v0.02](https://cdn.jsdelivr.net/gh/robinmin/imglanding/images/202303052318479.gif)



My trick is to specify the chat history path to a subfolder under my [Obsidian](https://obsidian.md/) data folder via the `-l` parameter. then I can use this brilliant tool to manage the conversation history. Going forward, It will be one of the next move to do more NLP-related in-depth development in this direction.

![image-20230305144703427](https://cdn.jsdelivr.net/gh/robinmin/imglanding/images/202303051447652.png)



#### Version History
- v0.0.4 at 2023-03-06 : get github release and brew installation ready.
- v0.0.2 at 2023-03-05 : Add deepl translate/lookup function support.
- v0.0.1 at 2023-03-04 : Project Initialize.


#### Reffernce
- [openai.com API Docs](https://platform.openai.com/docs/introduction/overview)
- [How to publish your Go binary as Homebrew Formula with GoReleaser](https://franzramadhan.com/posts/8-how-to-publish-go-binary-to-homebrew/)
- [Create a Custom CLI Tool and Distribute with HomeBrew Using Goreleaser and Github Actions](https://askcloudarchitech.com/posts/tutorials/create-homebrew-tap-golang-goreleaser-cobra-cli/)
- [Making your project available through Homebrew](https://dev.to/superfola/making-your-project-available-through-homebrew-1ll5)
- [Goreleaser Quick Start](https://goreleaser.com/quick-start/)
