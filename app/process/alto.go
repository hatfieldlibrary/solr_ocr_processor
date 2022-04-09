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

func (processor AltoProcessor) ProcessOcr(uuid *string, fileName string, alto *[]byte, position int,
	manifestId string, settings model.Configuration, log *log.Logger) error {
	updatedOcr, err := updateAlto(alto, position, settings)
	if err != nil {
		return err
	}
	if settings.ConvertToMiniOcr {
		updatedOcr, err = convertToMiniOcr(updatedOcr, position, settings)
		if err != nil {
			return err
		}
	}
	if settings.IndexType == "lazy" {
		err = PostToSolrLazyLoad(uuid, fileName, updatedOcr, manifestId, settings, log)
		if err != nil {
			return errors.New("ALTO indexing failed: " + err.Error())
		}
	} else {
		err = PostToSolr(uuid, fileName, updatedOcr, manifestId, settings, log)
		if err != nil {
			return errors.New("ALTO indexing failed: " + err.Error())
		}
	}
	return nil
}

// updateAlto sets the Page identifier and if required by configuration coverts unicode
// characters.
func updateAlto(alto *[]byte, position int, settings model.Configuration) (*string, error) {

	// There is no need to date when full indexing without character conversion is requested.
	if !settings.EscapeUtf8 && settings.IndexType != "lazy" {
		out := string(*alto)
		return &out, nil
	}

	var buffer bytes.Buffer
	reader := bytes.NewReader(*alto)
	decoder := xml.NewDecoder(reader)
	encoder := xml.NewEncoder(&buffer)

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
		case xml.StartElement:
			if t.Name.Local == "Page" {
				id := "Page." + strconv.Itoa(position)
				pos := getPosition(t, "ID")
				t.Attr[pos].Value = id
				if err := encoder.EncodeToken(t); err != nil {
					return nil, err
				}
				continue
			}

			if t.Name.Local == "String" && settings.EscapeUtf8 && settings.IndexType == "lazy" {
				pos := getPosition(t, "CONTENT")
				t.Attr[pos].Value = ToXmlCodePoint(t.Attr[pos].Value)
				if err := encoder.EncodeToken(t); err != nil {
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
		log.Fatal(err)
	}

	out := buffer.String()
	updated := fixResponse(&out, settings)
	return updated, nil

}

// convertToMiniOcr creates miniOcr output from the ALTO input.
func convertToMiniOcr(original *string, position int, settings model.Configuration) (*string, error) {
	reader := strings.NewReader(*original)
	decoder := xml.NewDecoder(reader)

	ocr := &model.OcrEl{}

	pageElements := make([]model.P, 0)
	textBlockElements := make([]model.B, 0)
	lineElements := make([]model.L, 0)
	wordElements := make([]model.W, 0)

	escape := settings.EscapeUtf8 && settings.IndexType == "lazy"

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		switch t := token.(type) {

		case xml.StartElement:

			if t.Name.Local == "Page" {
				height := t.Attr[2].Value
				width := t.Attr[3].Value
				dims := width + " " + height
				textBlockElements = nil
				pageId := "Page." + strconv.Itoa(position)
				page := &model.P{Id: pageId, Dimensions: dims}
				pageElements = append(pageElements, *page)
				ocr.Pages = pageElements
				continue
			}
			if t.Name.Local == "ComposedBlock" {
				continue
			}
			if t.Name.Local == "TextBlock" {
				lineElements = nil
				block := &model.B{}
				lastPage := &ocr.Pages[len(ocr.Pages)-1]
				textBlockElements = append(textBlockElements, *block)
				lastPage.Blocks = textBlockElements
				continue
			}
			if t.Name.Local == "TextLine" {
				wordElements = nil
				lineBlock := &model.L{}
				lastPage := &ocr.Pages[len(ocr.Pages)-1]
				if len(textBlockElements) > 0 {
					lastBlock := &lastPage.Blocks[len(textBlockElements)-1]
					lineElements = append(lineElements, *lineBlock)
					lastBlock.Lines = lineElements
				}
				continue
			}
			if t.Name.Local == "String" {
				content := t.Attr[0]
				height := t.Attr[1]
				width := t.Attr[2]
				vpos := t.Attr[3]
				hpos := t.Attr[4]
				var str = ""
				if escape {
					str = ToXmlCodePoint(content.Value)
				} else {
					str = content.Value
				}

				if len(content.Value) > 0 {
					coordinates := hpos.Value + " " + vpos.Value + " " + width.Value + " " + height.Value
					wordElement := model.W{CoorinateAttr: coordinates, Content: str + " "}
					lastPage := &ocr.Pages[len(ocr.Pages)-1]
					lastBlock := &lastPage.Blocks[len(textBlockElements)-1]
					currentLine := &lastBlock.Lines[len(lineElements)-1]
					wordElements = append(wordElements, wordElement)
					currentLine.Words = wordElements
				}
				continue
			}

		}

	}
	marshalledXml, err := xml.Marshal(ocr)
	if err != nil {
		return nil, err
	}
	out := string(marshalledXml)
	if settings.IndexType == "full" {
		// use single quotes to submit the XML in solr post
		out = strings.ReplaceAll(out, "\"", "'")
	}

	return &out, nil

}
