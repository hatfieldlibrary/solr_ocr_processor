package process

import "encoding/xml"

func getPosition(elem xml.StartElement, str string) int {
	for i := range elem.Attr {
		if elem.Attr[i].Name.Local == str {
			return i
		}
	}
	return -1
}

func hasClassValue(elem xml.StartElement, str string) bool {
	for i := range elem.Attr {
		if elem.Attr[i].Value == str {
			return true
		}
	}
	return false
}
