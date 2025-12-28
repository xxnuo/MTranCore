package MTranCore

import (
	"fmt"
	"strings"

	emoji "github.com/Andrew-M-C/go.emoji"
)

type emojiReplacement struct {
	original    string
	placeholder string
}

func processTextWithEmojiHandling(text string, processFunc func(string) (string, error)) (string, error) {
	var replacements []emojiReplacement
	var builder strings.Builder
	builder.Grow(len(text))

	for it := emoji.IterateChars(text); it.Next(); {
		if it.CurrentIsEmoji() {
			placeholder := fmt.Sprintf("\x00E%d\x00", len(replacements))
			replacements = append(replacements, emojiReplacement{
				original:    it.Current(),
				placeholder: placeholder,
			})
			builder.WriteString(placeholder)
		} else {
			builder.WriteString(it.Current())
		}
	}

	processedText, err := processFunc(builder.String())
	if err != nil {
		return "", err
	}

	result := processedText
	for _, r := range replacements {
		result = strings.ReplaceAll(result, r.placeholder, r.original)
	}

	return result, nil
}
