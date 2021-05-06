package internal

import "strconv"

func indexFiles(uuid string, annotationsMap map[string]string, altoFiles []string,
	manifestId string, settings Configuration) {
	for i := 0; i < len(altoFiles); i++ {
		if len(altoFiles[i]) > 0 {
			alto := getAltoXml(annotationsMap[altoFiles[i]])
			escapedAlto := escapeAlto(alto)
			postToSolr(uuid, altoFiles[i], escapedAlto, manifestId, "identifier", settings)
		}
	}
}

func escapeAlto(alto string) string {
	escapedAlto := ""
	for _, runeValue := range alto {
		if runeValue > 127 {
			escapedAlto += convertRune(runeValue)
		} else {
			escapedAlto += string(runeValue)
		}
	}
	return escapedAlto
}

func convertRune(rune rune) string {
	newValue := strconv.FormatInt(int64(rune), 16)
	ref := "&#" + newValue +";"
	return ref
}
