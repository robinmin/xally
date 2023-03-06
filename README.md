# X-Ally

(v0.0.4)

X-Ally is an AI-based TUI (aka Terminal User Interface) tool that helps people do things more elegantly. So far it has been integrated with the APIs of openai and deepl.


#### Usage

Before run the program, use your real keys to setup the environment variables as shown below:

```bash
# key from openai.com, mandatory so far
export OPENAI_API_KEY=sk-XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
# key from deepl.com, optional
export DEEPL_API_KEY=XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX

# run the program
$ ./build/bin/xally
```
If you want to know the details, please use the following command:
```bash
./build/bin/xally -h
```
then you will get the following tips:
```bash
xally version: xally/0.0.2
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



#### Installation

```bash
#$ go install github.com/robinmin/xally@latest
$ cd xally
$ make build
```
> mkdir -p build/bin
> GO111MODULE=on go build -o build/bin/xally ./cmd/client/main.go
> chmod u+x build/bin/xally

#### Version History
- v0.0.2 at 2023-03-05 : Add deepl translate/lookup function support
- v0.0.1 at 2023-03-04 : Project Initialize


#### Reffernce
- [openai.com API Docs](https://platform.openai.com/docs/introduction/overview)
