package MTranCore

import (
	"strings"
	"testing"
)

func TestProcessTextWithEmojiHandling(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"plain text", "hello world", "HELLO WORLD"},
		{"single emoji", "hello 😀 world", "HELLO 😀 WORLD"},
		{"multiple emojis", "🎉 party 🎊 time 🎈", "🎉 PARTY 🎊 TIME 🎈"},
		{"emoji at start", "🚀 launch", "🚀 LAUNCH"},
		{"emoji at end", "done ✅", "DONE ✅"},
		{"consecutive emojis", "👍👍👍", "👍👍👍"},
		{"flag emoji", "China 🇨🇳 Japan 🇯🇵", "CHINA 🇨🇳 JAPAN 🇯🇵"},
		{"skin tone", "thumbs up 👍🏻👍🏿", "THUMBS UP 👍🏻👍🏿"},
		{"ZWJ family", "family 👨‍👩‍👧‍👦 emoji", "FAMILY 👨‍👩‍👧‍👦 EMOJI"},
		{"rainbow flag", "pride 🏳️‍🌈 flag", "PRIDE 🏳️‍🌈 FLAG"},
		{"empty string", "", ""},
		{"only emoji", "🔥", "🔥"},
		{"mixed symbols", "© ® ™", "© ® ™"},
	}

	upperFunc := func(s string) (string, error) {
		return strings.ToUpper(s), nil
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := processTextWithEmojiHandling(tt.input, upperFunc)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestProcessTextWithEmojiHandling_PreservesOrder(t *testing.T) {
	input := "a🔴b🟢c🔵d"
	expected := "A🔴B🟢C🔵D"

	result, err := processTextWithEmojiHandling(input, func(s string) (string, error) {
		return strings.ToUpper(s), nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestProcessTextWithEmojiHandling_EmojiRestoration(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"simple emoji", "test 😀 message"},
		{"flag sequence", "🇺🇸 USA 🇨🇳 China"},
		{"ZWJ sequence", "👨‍👩‍👧‍👦 family"},
		{"skin tone modifier", "👋🏻👋🏿"},
		{"keycap sequence", "Press 1️⃣ then 2️⃣"},
		{"rainbow flag", "🏳️‍🌈 pride"},
		{"mixed complex", "Hello 👨‍💻 coding 🏳️‍🌈 pride 🇯🇵 japan"},
	}

	identityFunc := func(s string) (string, error) {
		return s, nil
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := processTextWithEmojiHandling(tt.input, identityFunc)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.input {
				t.Errorf("emoji not restored correctly\ngot:  %q\nwant: %q", result, tt.input)
				t.Errorf("got bytes:  %v", []byte(result))
				t.Errorf("want bytes: %v", []byte(tt.input))
			}
		})
	}
}

func TestProcessTextWithEmojiHandling_PlaceholderCollision(t *testing.T) {
	input := "[E0] 😀 [E1] 🎉 real text"

	result, err := processTextWithEmojiHandling(input, func(s string) (string, error) {
		return s, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != input {
		t.Errorf("placeholder collision detected\ngot:  %q\nwant: %q", result, input)
	}
}

func TestProcessTextWithEmojiHandling_TransformPreservesEmoji(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		transform func(string) (string, error)
		expected  string
	}{
		{
			"reverse words",
			"hello 😀 world",
			func(s string) (string, error) {
				words := strings.Fields(s)
				for i, j := 0, len(words)-1; i < j; i, j = i+1, j-1 {
					words[i], words[j] = words[j], words[i]
				}
				return strings.Join(words, " "), nil
			},
			"world 😀 hello",
		},
		{
			"add prefix suffix",
			"🚀 launch",
			func(s string) (string, error) {
				return "[START]" + s + "[END]", nil
			},
			"[START]🚀 launch[END]",
		},
		{
			"duplicate text",
			"hi 👋",
			func(s string) (string, error) {
				return s + " " + s, nil
			},
			"hi 👋 hi 👋",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := processTextWithEmojiHandling(tt.input, tt.transform)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func BenchmarkProcessTextWithEmojiHandling(b *testing.B) {
	text := "Hello 👋 World 🌍! This is a test 🧪 with emojis 😀🎉🚀"
	upperFunc := func(s string) (string, error) {
		return strings.ToUpper(s), nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = processTextWithEmojiHandling(text, upperFunc)
	}
}

func BenchmarkProcessTextWithEmojiHandling_NoEmoji(b *testing.B) {
	text := "Hello World! This is a test without any emojis at all."
	upperFunc := func(s string) (string, error) {
		return strings.ToUpper(s), nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = processTextWithEmojiHandling(text, upperFunc)
	}
}

func BenchmarkProcessTextWithEmojiHandling_ManyEmojis(b *testing.B) {
	text := "🎉🎊🎈🎁🎀🎄🎃🎇🎆🧨✨🎍🎎🎏🎐🎑🧧🎫🎟️🎪🎭🎨🎬🎤🎧🎼🎹🥁🎷🎺🎸🪕🎻🎲🎯🎳🎮🕹️🎰"
	upperFunc := func(s string) (string, error) {
		return strings.ToUpper(s), nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = processTextWithEmojiHandling(text, upperFunc)
	}
}
