package process

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/mspalti/altoindexer/model"
	"io"
	"log"
	"strings"
)

func (processor MiniOcrProcessor) ProcessOcr(uuid *string, fileName string, ocr *string, position int,
	manifestId string, settings model.Configuration, log *log.Logger) error {

	if settings.EscapeUtf8 && settings.IndexType == "lazy" {
		var buffer bytes.Buffer
		reader := bytes.NewReader([]byte(*ocr))
		decoder := xml.NewDecoder(reader)
		encoder := xml.NewEncoder(&buffer)

		for {
			token, err := decoder.Token()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Printf("error getting token: %t\n", err)
				return err
				break
			}
			switch t := token.(type) {
			case xml.StartElement:
				if t.Name.Local == "w" {
					var word model.W
					err := decoder.DecodeElement(&word, &t)
					if err != nil {
						fmt.Println(err)
					}
					t.Attr = t.Attr[:0]
					word.Content = toXmlCodePoint(word.Content)
					if err = encoder.EncodeElement(word, t); err != nil {
						return err
					}
				}
				if t.Name.Local == "p" {
					if err = encoder.EncodeToken(t); err != nil {
						return err

					}
				}
				if t.Name.Local == "b" {
					if err = encoder.EncodeToken(t); err != nil {
						return err

					}
				}
				if t.Name.Local == "l" {
					if err = encoder.EncodeToken(t); err != nil {
						return err

					}
				}
				if t.Name.Local == "ocr" {
					if err = encoder.EncodeToken(t); err != nil {
						return err
					}
				}
			case xml.EndElement:
				if t.Name.Local == "p" {
					if err = encoder.EncodeToken(t); err != nil {
						return err
					}
				}
				if t.Name.Local == "b" {
					if err = encoder.EncodeToken(t); err != nil {
						return err
					}
				}
				if t.Name.Local == "l" {
					if err = encoder.EncodeToken(t); err != nil {
						return err
					}
				}
				if t.Name.Local == "ocr" {
					if err = encoder.EncodeToken(t); err != nil {
						return err
					}
				}
			}

		}
		if err := encoder.Flush(); err != nil {
			return err

		}

		out := buffer.String()
		if settings.IndexType == "full" {
			out = strings.ReplaceAll(out, "\"", "'")
		}

		err := submitToIndex(uuid, fileName, &out, manifestId, settings, log)
		if err != nil {
			return err
		}
		return nil
	}

	err := submitToIndex(uuid, fileName, ocr, manifestId, settings, log)
	if err != nil {
		return err
	}

	return nil
}

// submitToIndex add to solr index
func submitToIndex(uuid *string, fileName string, ocr *string, manifestId string,
	settings model.Configuration, log *log.Logger) error {
	if settings.IndexType == "full" {
		var err = postToSolr(uuid, fileName, ocr, manifestId, settings, log)
		if err != nil {
			return errors.New("solr indexing failed: " + err.Error())
		}
	} else {
		var err = postToSolrLazyLoad(uuid, fileName, ocr, manifestId, settings, log)
		if err != nil {
			return errors.New("solr indexing failed: " + err.Error())
		}
	}
	return nil
}
