package pomodoro

import (
	"encoding/json"
	"fmt"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

func init() {
	initI18n()
}

type i18nTemplate struct {
	msg map[string]any
}

var (
	I18nBundle *i18n.Bundle
)

func initI18n() {
	I18nBundle = i18n.NewBundle(language.Japanese)
	I18nBundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	parse := func(lang language.Tag, t i18nTemplate) {
		if msg, err := json.Marshal(t.msg); err == nil {
			I18nBundle.MustParseMessageFileBytes(msg, lang.String()+".json")
		}
	}

	I18nTemplates := map[language.Tag]i18nTemplate{
		language.English: {
			map[string]any{
				// task
				"Start your task!": "Start your task!",
				"Task will end in Min minutes.": map[string]interface{}{
					"one":   "The break will end in {{ .Min }} minute.",
					"other": "The break will end in {{ .Min }} minutes.",
				},
				"The task will end at DateTime.": "The task will end at DateTime.",
				// break
				"The break has started!": "Pomodoro break time has started!",
				"The break will end in Min minutes.": map[string]interface{}{
					"one":   "The break will end in {{ .Min }} minute.",
					"other": "The break will end in {{ .Min }} minutes.",
				},
				"The break will end at DateTime.": "The break will end at DateTime.",
			},
		},
		language.Japanese: {
			map[string]any{
				// task
				"Start your task!":               "タスク開始なのん! ╲(๑˙Δ˙๑)",
				"Task will end in Min minutes.":  "タスクは{{ .Min }}分間ですん!",
				"The task will end at DateTime.": "{{ .DateTime }} まで頑張るのん! https://i.gyazo.com/c5e4b87ff10276499a02c164ea58dc96.png",
				// break
				"The break has started!":             "休憩なのん ฅ(๑¯Δ¯๑)",
				"The break will end in Min minutes.": "時間は五分間しかないのんな _(　　_‾ω‾ )_ぐでーん",
				"The break will end at DateTime.":    "時間は {{ .DateTime }} までなのん c⌒っ_ω_)っぐてーん https://i.gyazo.com/400a0d826b71bccbabf8e92236ef5b4f.png",
			},
		},
	}

	for lang, t := range I18nTemplates {
		parse(lang, t)
	}

	assertI18nTemplateMissing := func(ts map[language.Tag]i18nTemplate) {
		if len(ts) == 0 {
			return
		}
		langs := map[language.Tag]bool{}

		var t1 i18nTemplate
		var lang1 language.Tag
		i := 0
		for lang2, t2 := range ts {
			if i == 0 {
				t1 = t2
				lang1 = lang2
				langs[lang2] = true
				i++
				continue
			}
			i++

			// lang must be unique
			if _, ok := langs[lang2]; ok {
				panic(fmt.Sprintf("duplicate lang: %s", lang2))
			}
			langs[lang2] = true

			// key check
			if len(t1.msg) != len(t2.msg) {
				panic(fmt.Sprintf("%s and %s have different number of messages", lang1, lang2))
			}
			for k := range t1.msg {
				if _, ok := t2.msg[k]; !ok {
					panic(fmt.Sprintf("%s is missing %s", lang2, k))
				}
			}
		}
	}
	assertI18nTemplateMissing(I18nTemplates)

}
