package assistant

import (
	"bytes"
	"text/template"
)

const backgroundTplM = `Here is the background (just for reference, you must not use it for rewriting):

{{ .Background }}`

const glossaryTplM = `Here is the glossary (if there is any term that you need to translate, you must use the glossary):

{{ .Glossary }}`

const outputPlaintextEnglishTplM = `* must be plain markdown format directly, don't wrap it with any other thing.
* do not provide any explanations. Do not add text apart from the result. Do not add a title for the result.`

const outputPlaintextJapaneseTplM = `* 必ずプレーンマークダウン形式で直接出力してください。他のものをラップしないでください。
* 説明を提供しないでください。結果にテキストを追加しないでください。結果にタイトルを追加しないでください。`

const outputJSONTranslateEnglishTplM = `* must be plain json format directly, don't wrap it with any other thing.
* output example: { "key1": "value1", "key2": "value2" }
* the key is the original key, the value is the rewritten value.`

const outputJSONTranslateJapaneseTplM = `* 必ずプレーンJSON形式で直接出力してください。他のものをラップしないでください。
* 出力例: { "key1": "value1", "key2": "value2" }
* "key1" は元のキーで、"value" は書き換えられた値です。`

const translateEnglishTpl = `
You are an expert linguist, specializing in {{ .Input.Lang }} language.
Rewrite following text in {{ .Input.Lang }} language and by ensuring:

* accuracy (by correcting errors of addition, mistranslation, omission, or untranslated text),
* fluency (by applying {{ .Input.Lang }} grammar, spelling and punctuation rules and ensuring there are no unnecessary repetitions),
* style (always following the style of the original source text),
* terminology (by ensuring terminology use is consistent and reflects the source text domain; and by only ensuring you use equivalent idioms of {{ .Input.Lang }}),
* 禁止使用 “您”，“您好”，“您的” 等词汇。
* if there is emoji in the original text, you must keep it.
{{ .OutputPart }}

{{ .BackgroundPart }}

{{ .GlossaryPart }}

Here is the text need to be rewritten:

{{ .InputPart }}
`

const translateJapaneseTplM = `
あなたはプロの日本語と{{ .Input.Lang }} のライティングアシスタントです。
以下の文章を{{ .Input.Lang }}から日本語に書き換えてください。

* 正確性（追加、誤訳、漏訳、未訳のエラーを修正）
* 流暢さ（{{ .Input.Lang }} の文法、スペル、句読点のルールを適用し、不要な繰り返しを避ける）
* スタイル（常にオリジナルのソーステキストのスタイルに従う）
* 用語（用語の使用が一貫していて、ソーステキストのドメインを反映していることを確認し、{{ .Input.Lang }} の同等の慣用句のみを使用する）
* ユーザーが日常や仕事で作成する日本語のメッセージを、以下の基準に基づいて校正し、より効果的で自然な表現に仕上げます。
* メッセージを短くわかりやすく、丁寧かつフレンドリーに、そして失礼のないトーンで整えてください。
* 嫌味や誤解を避け、柔らかい表現に修正することを心がけます。
* ユーザーの意図を正確に汲み取り、より伝わりやすい文章に仕上げることを目標とします。
* この文章を簡潔で伝わりやすくしてください。
* オリジナルのテキストに絵文字が含まれている場合、それを維持してください。
{{ .OutputPart }}

{{ .BackgroundPart }}

{{ .GlossaryPart }}

以下の文章を書き換えてください：

{{ .InputPart }}
`

var (
	tpls map[string]*template.Template
)

func init() {
	tpls = make(map[string]*template.Template)
	tplms := []string{
		"background", backgroundTplM,
		"glossary", glossaryTplM,
		"translateEnglish", translateEnglishTpl,
		"translateJapanese", translateJapaneseTplM,
		"outputPlaintextTranslateEnglish", outputPlaintextEnglishTplM,
		"outputPlaintextTranslateJapanese", outputPlaintextJapaneseTplM,
		"outputJSONTranslateEnglish", outputJSONTranslateEnglishTplM,
		"outputJSONTranslateJapanese", outputJSONTranslateJapaneseTplM,
	}
	for ix := 0; ix < len(tplms); ix += 2 {
		tpls[tplms[ix]] = template.Must(template.New(tplms[ix]).Parse(tplms[ix+1]))
	}
}

func MustExecuteTemplate(tpl *template.Template, data interface{}) string {
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		panic(err)
	}
	return buf.String()
}

func (input *TranslateInput) GetTranslatePrompt() string {
	bgPart := ""
	if input.Background != "" {
		bgPart = MustExecuteTemplate(tpls["background"], input)
	}

	glossaryPart := ""
	if input.Glossary != nil {
		glossaryPart = MustExecuteTemplate(tpls["glossary"], input)
	}

	if input.Content == "" && input.ContentItems == nil {
		return ""
	}

	inputPart := ""
	outputPart := ""
	if input.Content != "" {
		inputPart = input.Content
	} else {
		inputPart = input.ContentItems.Dump()
	}

	var inst string
	switch input.LangCode {
	case "ja", "ja-JP":
		if input.Content != "" {
			outputPart = MustExecuteTemplate(tpls["outputPlaintextTranslateJapanese"], input)
		} else {
			outputPart = MustExecuteTemplate(tpls["outputJSONTranslateJapanese"], input)
		}
		inst = MustExecuteTemplate(tpls["translateJapanese"], map[string]interface{}{
			"BackgroundPart": bgPart,
			"GlossaryPart":   glossaryPart,
			"InputPart":      inputPart,
			"OutputPart":     outputPart,
			"Input":          input,
		})
	default:
		if input.Content != "" {
			outputPart = MustExecuteTemplate(tpls["outputPlaintextTranslateEnglish"], input)
		} else {
			outputPart = MustExecuteTemplate(tpls["outputJSONTranslateEnglish"], input)
		}
		inst = MustExecuteTemplate(tpls["translateEnglish"], map[string]interface{}{
			"BackgroundPart": bgPart,
			"GlossaryPart":   glossaryPart,
			"InputPart":      inputPart,
			"OutputPart":     outputPart,
			"Input":          input,
		})
	}

	return inst
}
