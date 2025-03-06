# translate-cli

A command-line interface (CLI) tool that uses AI to translate locale files based on JSON format.

## Features

- support multiple AI providers
- support glossary
- support adding background information
- batch translation
- no "您" or "您好" or "您的" in Chinese
- improved writing style in Japanese

## Usage

```bash
# create empty locale files
$ echo "{}" > example/langs/ja.json
$ echo "{}" > example/langs/zh-TW.json

# translate
$ translate-cli translate -s example/langs/en-US.json -d example/langs -g example/glossary.json -b example/background.txt --batch=20
📦 batch size: 20
📄 source:
  - file: example/langs/en-US.json
  - records: 12
📖 glossary:
  - file: example/glossary.json
📚 background:
  - file: example/background.txt
🌍 translating ...
✅ example/langs/ja.json: 12/12
✅ example/langs/zh-TW.json: 12/12
```

in which,

- `-s`: the source locale file to be used as a reference for translation.
- `-d`: the directory where the locale files are located.
- `-g`: the glossary file to be used as a reference for translation.
- `--batch`: the batch size for translation. default is 5.
  - in batch mode, it will arrange `batch` size items into a JSON object and send it to the AI provider at once.
  - the larger size may increase the cost of tokens, but it may also improve the translation quality as well.
  - if the size is too large, it may cause out of context window error.
  - some AI providers issue to handle complex JSON format, if you encounter this issue, you can try to reduce the size to 1

## Install

Please check the latest release [here](https://github.com/quailyquaily/translate-cli/tags), and download the binary for your platform.

Extract the binary and put it in your `$PATH` environment variable.

## Install from source

```bash
go install github.com/quailyquaily/translate-cli@latest
```

## How it works

`translate-cli` reads the directory containing locale files, retrieves translations from AI, and then writes the translated content back to the same files.

To have `translate-cli` translate the content of a JSON locale file, any existing values will be ignored.

If you want `translate-cli` to translate a specific value, you can add a "!" at the beginning of the string. Alternatively, you can delete the key/value pair from the JSON file to have `translate-cli` generate a new translation.

## Supported AI providers

- [x] OpenAI
- [x] OpenAI compatible API (e.g. DeepSeek, Grok)
- [ ] Azure OpenAI
- [ ] Bedrock
- [ ] Susano

## Config

the default config file is `~/.config/translate-cli/config.yaml`.

example config file:

```yaml
debug: false
openai:
  api_key: "sk-..."
  api_base: https://api.openai.com/v1
  model: "gpt-4o-mini"
provider: openai
```

in which,

- `debug`: This flag specifies whether to enable debug mode.
- `openai`: This section specifies the OpenAI compatible API configuration.
  - `api_key`: This flag specifies the API key for the OpenAI compatible API.
  - `api_base`: This flag specifies the base URL for the OpenAI compatible API.
  - `model`: This flag specifies the model to be used for the OpenAI compatible API.
- `susano`: This section specifies the Susano API configuration.
  - `api_key`: This flag specifies the API key for the Susano API.
  - `api_base`: This flag specifies the endpoint for the Susano API.
- `bedrock`: This section specifies the Bedrock API configuration.
  - `key`: This flag specifies the key for the Bedrock API.
  - `secret`: This flag specifies the secret for the Bedrock API.
  - `model`: This flag specifies the ARN of the model to be used for the Bedrock API.
- `azure`: This section specifies the Azure OpenAI API configuration.
  - `api_key`: This flag specifies the API key for the Azure OpenAI API.
  - `endpoint`: This flag specifies the endpoint for the Azure OpenAI API.
  - `model`: This flag specifies the model to be used for the Azure OpenAI API.
- `provider`: This flag specifies the AI provider. possible values are:
  - `openai`: use openai compacible API
  - `susano`: use susano api
  - `bedrock`: use bedrock api
  - `azure`: use azure openai api
