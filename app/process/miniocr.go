package process

import (
	"bytes"
	"encoding/xml"
	"errors"
	"github.com/mspalti/ocrprocessor/model"
	"io"
	"log"
	"strconv"
	"strings"
)

func (processor MiniOcrProcessor) ProcessOcr(uuid *string, fileName string, ocr *string, position int,
	manifestId string, settings model.Configuration, log *log.Logger) error {

	miniOcr, err := updateXml(ocr, position, settings)
	if err != nil {
		return err
	}

	if settings.IndexType == "full" {
		var err = PostToSolr(uuid, fileName, miniOcr, manifestId, settings, log)
		if err != nil {
			return errors.New("MiniOcr indexing failed: " + err.Error())
		}
	} else {
		var err = PostToSolrLazyLoad(uuid, fileName, miniOcr, manifestId, settings, log)
		if err != nil {
			return errors.New("MiniOcr indexing failed: " + err.Error())
		}
	}

	return nil
}

// updateXML updates the page ID and converts unicode to XML-encoded codepoint, if required by configuration.
func updateXml(ocr *string, position int, settings model.Configuration) (*string, error) {
	var buffer bytes.Buffer
	reader := bytes.NewReader([]byte(*ocr))
	decoder := xml.NewDecoder(reader)
	encoder := xml.NewEncoder(&buffer)

	xmlEncodeWord := false

	for {
		token, err := decoder.RawToken()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("error getting token: %t\n", err)
			return nil, err
		}
		switch t := token.(type) {
		case xml.CharData:
			if xmlEncodeWord {
				escaped := []byte(ToXmlCodePoint(string(t)))
				t = escaped
				if err = encoder.EncodeToken(t); err != nil {
					return nil, err

				}
				xmlEncodeWord = false
				continue
			}
		case xml.StartElement:
			if t.Name.Local == "w" {
				if settings.EscapeUtf8 && settings.IndexType == "lazy" {
					xmlEncodeWord = true
				}
				if err = encoder.EncodeToken(t); err != nil {
					return nil, err
				}
				continue
			}
			if t.Name.Local == "p" {
				pos := getPosition(t, "id")
				t.Attr[pos].Value = "Page." + strconv.Itoa(position)
				if err = encoder.EncodeToken(t); err != nil {
					return nil, err
				}
				continue
			}
		}

		if err := encoder.EncodeToken(xml.CopyToken(token)); err != nil {
			return nil, err
		}

	}
	if err := encoder.Flush(); err != nil {
		return nil, err

	}

	out := buffer.String()
	out = strings.ReplaceAll(out, "\n", "")
	if settings.IndexType == "full" {
		out = strings.ReplaceAll(out, "\"", "'")
	}

	return &out, nil

}
