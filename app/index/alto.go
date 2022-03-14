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
				updatedAlto, err := updateAlto(&altoString, i, settings.EscapeUtf8)
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

func updateAlto(alto *string, position int, escapeUtf8 bool) (*string, error) {
	var buffer bytes.Buffer
	reader := strings.NewReader(*alto)
	decoder := xml.NewDecoder(reader)
	encoder := xml.NewEncoder(&buffer)
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
			if t.Name.Local == "Page" {
				var page Page
				if err = decoder.DecodeElement(&page, &t); err != nil {
					log.Fatal(err)
				}

				// modify the version value and encode the element back
				page.Id = "Page." + strconv.Itoa(position)

				t.Attr = t.Attr[:0]
				if err = encoder.EncodeElement(page, t); err != nil {
					log.Fatal(err)
				}
				continue
			}
			if t.Name.Local == "String" {
				var content String
				if err = decoder.DecodeElement(&content, &t); err != nil {
					log.Fatal(err)
				}
				if escapeUtf8 {
					content.CONTENT = toXmlCodePoint(content.CONTENT)
				}
				if err = encoder.EncodeElement(content, t); err != nil {
					log.Fatal(err)
				}
				continue
			}
			if err := encoder.EncodeToken(xml.CopyToken(t)); err != nil {
				log.Fatal(err)
			}
		}
	}

	// must call flush, otherwise some elements will be missing
	if err := encoder.Flush(); err != nil {
		log.Fatal(err)
	}
	out := buffer.String()
	return &out, nil

}
