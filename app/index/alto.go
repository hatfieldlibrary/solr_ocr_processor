package index

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
)

func indexFiles(uuid string, annotationsMap map[string]string, altoFiles []string,
	manifestId string, settings Configuration) error {
	for i := 0; i < len(altoFiles); i++ {
		if len(altoFiles[i]) > 0 {
			alto, err := getAltoXml(annotationsMap[altoFiles[i]])
			if err != nil {
				return errors.New("could not retrieve alto file from dspace")
			}
			if len(alto) != 0 {
				updatedAlto, err := setAltoId(&alto, i)
				if err != nil {
					return err
				}
				escapedAlto := escapeAlto(updatedAlto)
				err = postToSolr(uuid, altoFiles[i], escapedAlto, manifestId, settings)
				if err != nil {
					return errors.New("solr indexing failed: " + err.Error())
				}
			}
		}
	}
	return nil
}

func escapeAlto(alto *string) *string {

	var sb strings.Builder
	for _, runeValue := range *alto {
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
	ref := "&#" + newValue +";"
	return ref
}

func setAltoId(alto *string, position int) (*string, error) {
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
			if t.Name.Local == "Description"{
				continue
			}
			if t.Name.Local == "Styles"{
				continue
			}
			if t.Name.Local == "MeasurementUnit"{
				continue
			}
			if t.Name.Local == "fileName"{
				continue
			}
			if t.Name.Local == "sourceImageInformation"{
				continue
			}
			if t.Name.Local == "OCRProcessing"{
				continue
			}
			if t.Name.Local == "ocrProcessingStep"{
				continue
			}
			if t.Name.Local == "processingDateTime"{
				continue
			}
			if t.Name.Local == "processingSoftware"{
				continue
			}
			if t.Name.Local == "softwareCreator"{
				continue
			}
			if t.Name.Local == "softwareName"{
				continue
			}
			if t.Name.Local == "softwareVersion"{
				continue
			}
			if t.Name.Local == "ParagraphStyle"{
				continue
			}
			if err := encoder.EncodeToken(xml.CopyToken(t)); err != nil {
					log.Fatal(err)
			}

		case xml.EndElement:
			if t.Name.Local == "Description"{
				continue
			}
			if t.Name.Local == "Styles"{
				continue
			}
			if t.Name.Local == "MeasurementUnit"{
				continue
			}
			if t.Name.Local == "fileName"{
				continue
			}
			if t.Name.Local == "sourceImageInformation"{
				continue
			}
			if t.Name.Local == "OCRProcessing"{
				continue
			}
			if t.Name.Local == "ocrProcessingStep"{
				continue
			}
			if t.Name.Local == "processingDateTime"{
				continue
			}
			if t.Name.Local == "processingSoftware"{
				continue
			}
			if t.Name.Local == "softwareCreator"{
				continue
			}
			if t.Name.Local == "softwareName"{
				continue
			}
			if t.Name.Local == "softwareVersion"{
				continue
			}
			if t.Name.Local == "ParagraphStyle"{
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

