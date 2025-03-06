package translate

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/lyricat/goutils/ai"
	"github.com/lyricat/goutils/structs"
	"github.com/quailyquaily/translate-cli/cmd/parser"
	"github.com/quailyquaily/translate-cli/internal/assistant"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
)

var (
	dir            string
	sourceFile     string
	glossaryFile   string
	backgroundFile string
	batchSize      int
)

func NewCmd() *cobra.Command {
	translateCmd := &cobra.Command{
		Use: "translate",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			aiInst := ai.New(ai.Config{
				Provider:      viper.GetString("provider"),
				OpenAIAPIKey:  viper.GetString("openai.api_key"),
				OpenAIAPIBase: viper.GetString("openai.api_base"),
				OpenAIModel:   viper.GetString("openai.model"),
				Debug:         viper.GetBool("debug"),
			})

			ant := assistant.New(assistant.Config{
				Provider: viper.GetString("provider"),
			}, aiInst)

			source, others, glossary, background, err := provideFiles()
			if err != nil {
				cmd.PrintErrln(err)
				return
			}
			if batchSize > 1 {
				cmd.Println("📦 batch size:", batchSize)
			}
			cmd.Printf("📄 source:\n  - file: %s\n  - records: %d\n", source.Path, len(source.LocaleItemsMap))

			if glossary != nil {
				cmd.Printf("📖 glossary:\n  - file: %s\n", glossaryFile)
			}

			if background != "" {
				cmd.Printf("📚 background:\n  - file: %s\n", backgroundFile)
			}

			cmd.Println("🌍 translating ...")
			for _, item := range others {
				err = process(ctx, ant, source, item, glossary, background)
				if err != nil {
					cmd.PrintErrln("process failed: ", err)
					return
				}
			}
		},
	}

	translateCmd.Flags().StringVarP(&dir, "dir", "d", "", "the directory of language files")
	translateCmd.Flags().StringVarP(&sourceFile, "source", "s", "", "the source language file")
	translateCmd.Flags().StringVarP(&glossaryFile, "glossary", "g", "", "the glossary file")
	translateCmd.Flags().StringVarP(&backgroundFile, "background", "b", "", "the background file")
	translateCmd.Flags().IntVar(&batchSize, "batch", 5, "the batch size")

	return translateCmd
}

func process(ctx context.Context, ant *assistant.Assistant,
	source *parser.LocaleFileContent, target *parser.LocaleFileContent, glossary *parser.GlossaryContent, background string) error {
	count := 1

	groupResult := []*assistant.TranslateInput{}

	var localeItems []structs.JSONMap
	if batchSize > 1 {
		localeItems = source.LocaleItemsMap.Split(batchSize)
	} else {
		localeItems = []structs.JSONMap{source.LocaleItemsMap}
	}

	for _, items := range localeItems {
		// translate each group
		contentItems := structs.JSONMap{}
		for k, _v := range items {
			needToTranslate := false
			v := _v.(string)
			if v != "" {
				if _, ok := target.LocaleItemsMap[k]; !ok {
					// key does not exist, translate it
					needToTranslate = true
				} else {
					// key exists
					if target.LocaleItemsMap.GetString(k) == "" {
						// empty string, translate it
						needToTranslate = true
					} else if target.LocaleItemsMap.GetString(k)[0] == '!' {
						// value starts with "!", translate it
						needToTranslate = true
					}
				}

				if needToTranslate {
					if batchSize > 1 {
						contentItems[k] = v
					} else {
						groupResult = append(groupResult, &assistant.TranslateInput{
							Key:        k,
							Content:    v,
							Lang:       target.Lang,
							LangCode:   target.Code,
							Background: background,
							Glossary:   glossary.GetMapByLang(target.Code),
						})
					}
				}

				fmt.Printf("\r🔄 %s: %d/%d", target.Path, count, len(source.LocaleItemsMap))
				count += 1
			}
		}
		if batchSize > 1 {
			if len(contentItems) > 0 {
				groupResult = append(groupResult, &assistant.TranslateInput{
					ContentItems: contentItems,
					Lang:         target.Lang,
					LangCode:     target.Code,
					Background:   background,
					Glossary:     glossary.GetMapByLang(target.Code),
				})
			}
		}
	}

	if batchSize > 1 {
		for _, item := range groupResult {
			ret, err := ant.TranslateBatch(ctx, item)
			if err != nil {
				return err
			}
			for k, v := range ret {
				target.LocaleItemsMap.SetValue(k, v)
			}
		}
	} else {
		for _, item := range groupResult {
			result, err := ant.Translate(ctx, item)
			if err != nil {
				return err
			}
			target.LocaleItemsMap.SetValue(item.Key, result)
		}
	}

	buf, err := target.JSON()
	if err != nil {
		return err
	}

	err = os.WriteFile(target.Path, buf, 0644)
	if err != nil {
		return err
	}

	fmt.Printf("\r✅ %s: %d/%d\n", target.Path, len(source.LocaleItemsMap), len(source.LocaleItemsMap))

	return nil
}

func provideFiles() (source *parser.LocaleFileContent, others []*parser.LocaleFileContent, glossary *parser.GlossaryContent, background string, err error) {
	if glossaryFile != "" {
		glossary, err = parser.NewGlossaryFromJSONFile(glossaryFile)
		if err != nil {
			fmt.Printf("parse glossary file failed: %s. Use empty glossary.\n", err)
			glossary = &parser.GlossaryContent{
				Maps: make(map[string]*parser.GlossaryMapItem),
			}
		}
	}

	if backgroundFile != "" {
		buf, err := os.ReadFile(backgroundFile)
		if err != nil {
			fmt.Printf("read background file failed: %s. Use empty background.\n", err)
			background = ""
		} else {
			background = string(buf)
		}
	}

	if sourceFile != "" {
		source = &parser.LocaleFileContent{}
		if err = source.ParseFromJSONFile(sourceFile); err != nil {
			return
		}

		var lang string
		lang, err = langCodeToName("en-US")
		if err != nil {
			return
		}

		source.Code = "en-US"
		source.Lang = lang
	} else {
		err = fmt.Errorf("source file is required. use -s flag to specify the source file")
		return
	}

	if dir != "" {
		others = make([]*parser.LocaleFileContent, 0)
		items, _ := os.ReadDir(dir)
		sourceBaseFile := filepath.Base(sourceFile)
		for _, item := range items {
			if !item.IsDir() {
				name := filepath.Base(item.Name())
				ext := filepath.Ext(name)
				if strings.EqualFold(item.Name(), sourceBaseFile) {
					continue
				}

				if strings.ToLower(ext) != ".json" {
					fmt.Printf("file %s is not a JSON file. skip this file.\n", name)
					continue
				}

				localeContent := &parser.LocaleFileContent{}
				if err = localeContent.ParseFromJSONFile(path.Join(dir, item.Name())); err != nil {
					fmt.Printf("parse file failed: %s. Use default locale content.\n", err)
				}

				others = append(others, localeContent)
			}
		}
	} else {
		err = fmt.Errorf("dir is required. use -d flag to specify the directory of language files")
		return
	}

	return
}

func langCodeToName(code string) (string, error) {
	tag, err := language.Parse(code)
	if err != nil {
		return "", err
	}
	return display.Self.Name(tag), nil
}
