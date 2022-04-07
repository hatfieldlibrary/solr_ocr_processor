package process

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// ToXmlCodePoint converts non-ascii chars to xml code point, ignoring any non-utf8 encoding
func ToXmlCodePoint(str string) string {
	var sb strings.Builder
	for _, runeValue := range str {
		if runeValue > 127 {
			sb.WriteString(convertRune(runeValue))
		} else {
			sb.WriteString(string(runeValue))
		}
	}
	escapedStr := sb.String()
	return escapedStr

}

// convertRune converts rune to XML-escaped codepoint
func convertRune(rune rune) string {
	if rune == utf8.RuneError {
		return ""
	}
	codepoint := fmt.Sprint(rune)
	ref := "&#" + codepoint + ";"
	return ref
}
