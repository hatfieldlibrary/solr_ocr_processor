package process

import (
	"bytes"
	"encoding/xml"
	"errors"
	"github.com/mspalti/altoindexer/model"
	"io"
	"log"
	"strconv"
	"strings"
)

func (processor AltoProcessor) ProcessOcr(uuid *string, fileName string, alto *string, position int,
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
		err = postToSolrLazyLoad(uuid, fileName, updatedOcr, manifestId, settings, log)
		if err != nil {
			return errors.New("solr indexing failed: " + err.Error())
		}
	} else {
		err = postToSolr(uuid, fileName, updatedOcr, manifestId, settings, log)
		if err != nil {
			return errors.New("solr indexing failed: " + err.Error())
		}
	}
	return nil
}

func encodeStrings(strings []model.String) {
	for i, _ := range strings {
		strings[i].CONTENT = toXmlCodePoint(strings[i].CONTENT)
	}
}
func getTextLines(textLines []model.TextLine) {
	for i, _ := range textLines {
		encodeStrings(textLines[i].String)
	}
}
func getTextBlocks(textBlocks []model.TextBlock) {
	for i, _ := range textBlocks {
		getTextLines(textBlocks[i].TextLine)
	}
}
func getComposedBlocks(composedBlocks []model.ComposedBlock) {
	for i, _ := range composedBlocks {
		getTextBlocks(composedBlocks[i].TextBlock)
	}
}

func updateAlto(alto *string, position int, settings model.Configuration) (*string, error) {
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
				var model model.Alto
				if err = decoder.DecodeElement(&model, &t); err != nil {
					log.Fatal(err)
				}
				t.Attr = t.Attr[:0]
				model.Xmlns = ""
				model.Layout.Page.Id = "Page." + strconv.Itoa(position)
				if escape {
					getComposedBlocks(model.Layout.Page.PrintSpace.ComposedBlock)
					getTextBlocks(model.Layout.Page.PrintSpace.TextBlock)
				}
				if err = encoder.EncodeElement(model, t); err != nil {
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
			break
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
					str = toXmlCodePoint(content.Value)
				} else {
					str = content.Value
				}
				if len(str) > 0 {
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
	out = xml.Header + out
	if settings.IndexType == "full" {
		// use single quotes to submit the XML in solr post
		out = strings.ReplaceAll(out, "\"", "'")
	}

	return &out, nil

}
