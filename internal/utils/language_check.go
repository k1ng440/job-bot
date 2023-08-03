package utils

import (
	"strings"

	"github.com/pemistahl/lingua-go"
)

var detector = lingua.NewLanguageDetectorBuilder().
	FromAllLanguages().
	Build()

// DetectLanguage detects the language of the given text
func DetectLanguage(text string) string {
	lang, _ := detector.DetectLanguageOf(text)
	return strings.ToLower(lang.String())
}
