package index

import (
	"fmt"
	"strings"
)

func convertToAscii(input *string) *string {

	var sb strings.Builder
	for _, runeValue := range *input {
		if runeValue > 127 {
			sb.WriteString(convertRune(runeValue))
		} else {
			sb.WriteString(string(runeValue))
		}
	}
	escapedAlto := sb.String()
	return &escapedAlto
}

func convertRune(rune rune) string {
	newValue := fmt.Sprint(rune)
	ref := "&#" + newValue + ";"
	return ref
}
