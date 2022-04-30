package process

import (
	"encoding/xml"
	"github.com/mspalti/ocrprocessor/model"
	"strings"
)

// getPosition returns the position of the attribute in the token attribute list.
func getPosition(elem xml.StartElement, attribute string) int {
	for i := range elem.Attr {
		if elem.Attr[i].Name.Local == attribute {
			return i
		}
	}
	return -1
}

// hasClassValue return true if the class is found in the token attribute list
func hasClassValue(elem xml.StartElement, str string) bool {
	for i := range elem.Attr {
		if elem.Attr[i].Value == str {
			return true
		}
	}
	return false
}

// fixResponse converts double so single quotes and other cleanup when full indexing is requested.
// This utility function has no effect when Configuration requires subsequent conversion to
// MiniOcr format.
func fixResponse(input *string, settings model.Configuration) *string {
	if settings.IndexType == "full" && settings.ConvertToMiniOcr == false {
		tmp := strings.ReplaceAll(*input, "\n", "")
		tmp = strings.ReplaceAll(tmp, "\"", "'")
		return &tmp
	}
	return input
}

// getDSpaceApiEndpoint returns the URL for the DSpace IIIF endpoint
func getDSpaceApiEndpoint(host string, uuid string, iiiftype string) string {
	return host + "/iiif/" + uuid + "/" + iiiftype
}
