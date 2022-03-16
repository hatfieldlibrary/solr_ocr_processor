package index

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"log"
	"strconv"
	"strings"
)

func processAlto(uuid string, annotationsMap map[string]string, altoFiles []string,
	manifestId string, settings Configuration, log *log.Logger) error {
	for i := 0; i < len(altoFiles); i++ {
		if len(altoFiles[i]) > 0 {
			alto, err := getAltoXml(annotationsMap[altoFiles[i]], log)
			if err != nil {
				return errors.New("could not retrieve alto file from dspace")
			}
			if len(alto) != 0 {
				altoString := string(alto)

				updatedAlto, err := updateAlto(&altoString, i, settings)
				if err != nil {
					return err
				}
				if settings.IndexType == "lazy" {
					err = postToSolrLazyLoad(uuid, altoFiles[i], updatedAlto, manifestId, settings, log)
					if err != nil {
						return errors.New("solr indexing failed: " + err.Error())
					}
				} else {
					err = postToSolr(uuid, altoFiles[i], updatedAlto, manifestId, settings, log)
					if err != nil {
						return errors.New("solr indexing failed: " + err.Error())
					}
				}
			}
		}
	}
	return nil
}

func encodeStrings(strings []String) {
	for i, _ := range strings {
		strings[i].CONTENT = toXmlCodePoint(strings[i].CONTENT)
	}
}
func getTextLines(textLines []TextLine) {
	for i, _ := range textLines {
		encodeStrings(textLines[i].String)
	}
}
func getTextBlocks(textBlocks []TextBlock) {
	for i, _ := range textBlocks {
		getTextLines(textBlocks[i].TextLine)
	}
}
func getComposedBlocks(composedBlocks []ComposedBlock) {
	for i, _ := range composedBlocks {
		getTextBlocks(composedBlocks[i].TextBlock)
	}
}

func updateAlto(alto *string, position int, settings Configuration) (*string, error) {
	var buffer bytes.Buffer
	reader := strings.NewReader(*alto)
	decoder := xml.NewDecoder(reader)
	encoder := xml.NewEncoder(&buffer)

	escape := settings.EscapeUtf8 && settings.IndexType == "lazy"

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("error getting token: %t\n", err)
			return nil, err
			break
		}

		switch t := token.(type) {
		case xml.StartElement:
			if t.Name.Local == "alto" {
				var alto Alto
				if err = decoder.DecodeElement(&alto, &t); err != nil {
					log.Fatal(err)
				}
				t.Attr = t.Attr[:0]
				alto.Xmlns = ""
				alto.Layout.Page.Id = "Page." + strconv.Itoa(position)
				if escape {
					getComposedBlocks(alto.Layout.Page.PrintSpace.ComposedBlock)
					getTextBlocks(alto.Layout.Page.PrintSpace.TextBlock)
				}
				if err = encoder.EncodeElement(alto, t); err != nil {
					log.Fatal(err)
				}
				continue
			}
		}

	}

	// must call flush, otherwise some elements will be missing
	if err := encoder.Flush(); err != nil {
		log.Fatal(err)
	}

	out := buffer.String()
	if settings.IndexType == "full" {
		// Use single quotes in XML so that we submit in json. Note
		// that full indexing of ALTO is not advised. The more
		// compact miniocr format is preferred.
		out = strings.ReplaceAll(out, "\"", "'")
	}
	return &out, nil

}
