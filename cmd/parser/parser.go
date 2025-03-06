package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/lyricat/goutils/structs"
	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
)

type (
	LocaleFileContent struct {
		Code string
		Lang string
		Path string

		LocaleItemsMap structs.JSONMap
	}
	GlossaryContent struct {
		Maps map[string]*GlossaryMapItem
	}
	GlossaryMapItem map[string]string
)

func (gi *GlossaryMapItem) JSON() string {
	if gi == nil {
		return "{}"
	}

	buf, err := json.MarshalIndent(gi, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(buf)
}

func (gi *GlossaryMapItem) Set(key string, value string) {
	(*gi)[key] = value
}

func NewGlossaryFromJSONFile(path string) (*GlossaryContent, error) {
	var err error
	if _, err = os.Stat(path); err != nil {
		return nil, err
	}

	// Read the JSON file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Parse JSON into a temporary map
	var sourceMap map[string]map[string]string
	if err := json.Unmarshal(data, &sourceMap); err != nil {
		return nil, err
	}

	// Transform the structure
	result := &GlossaryContent{
		Maps: make(map[string]*GlossaryMapItem),
	}

	// Iterate through the source map and reorganize by language
	for term, translations := range sourceMap {
		for lang, translation := range translations {
			// Initialize the language map if needed
			if _, exists := result.Maps[lang]; !exists {
				mapItem := make(GlossaryMapItem)
				result.Maps[lang] = &mapItem
			}
			// Add the term -> translation mapping for this language
			result.Maps[lang].Set(term, translation)
		}
	}

	return result, nil
}

func (g *GlossaryContent) GetMapByLang(lang string) *GlossaryMapItem {
	if lang != "" && g.Maps != nil {
		if _, ok := g.Maps[lang]; !ok {
			return nil
		}
		return g.Maps[lang]
	}
	return nil
}

func (l *LocaleFileContent) ParseFromJSONFile(path string) error {
	var err error
	if _, err = os.Stat(path); err != nil {
		return err
	}

	name := filepath.Base(path) // get base name of file
	ext := filepath.Ext(name)   // get extension
	nameWithoutExt := name[0 : len(name)-len(ext)]

	if strings.ToLower(ext) != ".json" {
		return fmt.Errorf("file %s is not a json file", name)
	}

	lang, err := langCodeToName(nameWithoutExt)
	if err != nil {
		return err
	}

	l.Code = nameWithoutExt
	l.Lang = lang
	l.Path = path

	if l.LocaleItemsMap == nil {
		l.LocaleItemsMap = structs.NewJSONMap()
	}

	// read the json file
	sourceBytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// convert
	var data map[string]interface{}
	if err := json.Unmarshal(sourceBytes, &data); err != nil {
		l.LocaleItemsMap = structs.NewJSONMap()
		return nil
	}

	result := structs.NewJSONMap()
	flatten(data, result, "")

	l.LocaleItemsMap = result
	return nil
}

func (l *LocaleFileContent) JSON() ([]byte, error) {
	nestedData := nestedInsertion(l.LocaleItemsMap)
	sortedData := sortMapKeys(nestedData)

	jsonData, err := json.MarshalIndent(sortedData, "", "  ")
	if err != nil {
		return nil, err
	}
	return jsonData, nil
}

func flatten(input, result structs.JSONMap, currentKey string) {
	for key, value := range input {
		newKey := key
		if currentKey != "" {
			newKey = currentKey + "/" + key
		}
		switch child := value.(type) {
		case map[string]interface{}:
			flatten(child, result, newKey)
		default:
			result.SetValue(newKey, fmt.Sprint(value))
		}
	}
}

func nestedInsertion(input structs.JSONMap) map[string]interface{} {
	data := make(map[string]interface{})
	for key, value := range input {
		parts := strings.Split(key, "/")
		currentMap := data
		for i, part := range parts {
			if i == len(parts)-1 {
				currentMap[part] = value
			} else {
				if _, exist := currentMap[part]; !exist {
					currentMap[part] = make(map[string]interface{})
				}
				currentMap = currentMap[part].(map[string]interface{})
			}
		}
	}
	return data
}

func sortMapKeys(data interface{}) interface{} {
	switch data := data.(type) {
	case map[string]interface{}:
		keys := make([]string, 0, len(data))
		for key := range data {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		result := make(map[string]interface{}, len(data))
		for _, key := range keys {
			result[key] = sortMapKeys(data[key])
		}
		return result
	default:
		return data
	}
}

func langCodeToName(code string) (string, error) {
	tag, err := language.Parse(code)
	if err != nil {
		return "", err
	}
	return display.Self.Name(tag), nil
}
