package assistant

import (
	"context"
	"fmt"
	"time"

	"github.com/lyricat/goutils/structs"
	"github.com/quailyquaily/translate-cli/cmd/parser"
)

type (
	TranslateInput struct {
		// for single translate
		Key     string
		Content string
		// for batch translate
		ContentItems structs.JSONMap

		Lang       string                  // language name, e.g. "English", "Japanese"
		LangCode   string                  // language code, e.g. "en", "ja"
		Background string                  // the background of the translation
		Glossary   *parser.GlossaryMapItem // the glossary of the translation
	}
)

func (a *Assistant) Translate(ctx context.Context, input *TranslateInput) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	inst := input.GetTranslatePrompt()

	return a.AIRequestText(ctx, inst)
}

func (a *Assistant) TranslateBatch(ctx context.Context, input *TranslateInput) (structs.JSONMap, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	inst := input.GetTranslatePrompt()

	ret, err := a.AIRequestJSON(ctx, inst)
	if err != nil {
		return nil, err
	}

	// validate the result.
	// the size of ret.Json should be the same as the size of input.ContentItems
	if len(ret.Json) != len(input.ContentItems) {
		return nil, fmt.Errorf("the size of result is not the same as the size of input")
	}
	// and all the keys should be in the input.ContentItems
	for k := range ret.Json {
		if _, ok := input.ContentItems[k]; !ok {
			return nil, fmt.Errorf("the key %s is not in the input", k)
		}
	}

	return ret.Json, nil
}
