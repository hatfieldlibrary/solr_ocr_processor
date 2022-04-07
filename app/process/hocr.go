package process

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/mspalti/ocrprocessor/model"
	"io"
	"log"
	"regexp"
	"strconv"
	"strings"
)

var pageBBox = regexp.MustCompile(`bbox 0 0 (\d+) (\d+)`)
var wordBBox = regexp.MustCompile(`bbox (\d+) (\d+) (\d+) (\d+)`)

func (processor HocrProcessor) ProcessOcr(uuid *string, fileName string, ocr *string, position int,
	manifestId string, settings model.Configuration, log *log.Logger) error {
	updatedOcr, err := updateXML(ocr, position, settings)
	if err != nil {
		return err
	}
	if settings.ConvertToMiniOcr {
		updatedOcr, err = convert(updatedOcr, position, settings)
		if err != nil {
			return err
		}
	}
	if settings.IndexType == "lazy" {
		err = PostToSolrLazyLoad(uuid, fileName, updatedOcr, manifestId, settings, log)
		if err != nil {
			return errors.New("hOCR indexing failed: " + err.Error())
		}
	} else {
		err = PostToSolr(uuid, fileName, updatedOcr, manifestId, settings, log)
		if err != nil {
			return errors.New("hOCR indexing failed: " + err.Error())
		}
	}
	return nil

}

// convert returns MiniOcr for the original hOCR input.
func convert(original *string, position int, settings model.Configuration) (*string, error) {

	reader := strings.NewReader(*original)
	decoder := xml.NewDecoder(reader)

	ocr := &model.OcrEl{}

	pageElements := make([]model.P, 0)
	textBlockElements := make([]model.B, 0)
	lineElements := make([]model.L, 0)
	wordElements := make([]model.W, 0)

	isword := false
	var dims string

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
		case xml.CharData:
			if isword {
				lastPage := &ocr.Pages[len(ocr.Pages)-1]
				lastBlock := &lastPage.Blocks[len(textBlockElements)-1]
				currentLine := &lastBlock.Lines[len(lineElements)-1]
				wordElement := model.W{CoorinateAttr: dims, Content: " "}
				currentLine.Words = wordElements
				wordElement.Content = string(t) + " "
				wordElements = append(wordElements, wordElement)
				isword = false
				continue
			}
		case xml.StartElement:
			if hasClassValue(t, "ocr_page") {
				// p
				title := t.Attr[getPosition(t, "title")].Value
				bbox := pageBBox.FindSubmatch([]byte(title))
				var width string
				var height string
				if len(bbox) == 3 {
					width = string(bbox[1])
					height = string(bbox[2])
				}
				dims := width + " " + height
				pageId := "Page." + strconv.Itoa(position)
				page := &model.P{Id: pageId, Dimensions: dims}
				pageElements = append(pageElements, *page)
				ocr.Pages = pageElements
				continue
			}
			if hasClassValue(t, "ocr_carea") {
				// b
				lineElements = nil
				block := &model.B{}
				lastPage := &ocr.Pages[len(ocr.Pages)-1]
				textBlockElements = append(textBlockElements, *block)
				lastPage.Blocks = textBlockElements
				continue
			}
			if hasClassValue(t, "ocrx_block") {
				// b
				lineElements = nil
				block := &model.B{}
				lastPage := &ocr.Pages[len(ocr.Pages)-1]
				textBlockElements = append(textBlockElements, *block)
				lastPage.Blocks = textBlockElements
				continue
			}
			if hasClassValue(t, "ocr_par") {
				// not mapped
			}
			if hasClassValue(t, "ocr_line") {
				// l
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
			if hasClassValue(t, "ocrx_line") {
				// l
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
			if hasClassValue(t, "ocrx_word") {
				// w
				title := t.Attr[getPosition(t, "title")].Value
				bbox := wordBBox.FindSubmatch([]byte(title))
				var width string
				var height string
				var vpos string
				var hpos string
				if len(bbox) == 5 {
					hposInt, _ := strconv.ParseInt(string(bbox[1]), 10, 32)
					vposInt, _ := strconv.ParseInt(string(bbox[2]), 10, 32)
					hcorner, _ := strconv.ParseInt(string(bbox[3]), 10, 32)
					wcorner, _ := strconv.ParseInt(string(bbox[4]), 10, 32)
					hpos = string(bbox[1])
					vpos = string(bbox[2])
					width = strconv.FormatInt(hcorner-hposInt, 10)
					height = strconv.FormatInt(wcorner-vposInt, 10)

				}
				dims = fmt.Sprintf("%s %s %s %s", hpos, vpos, width, height)
				isword = true
			}
		}
	}
	marshalledXml, err := xml.Marshal(ocr)
	if err != nil {
		return nil, err
	}
	out := string(marshalledXml)
	if settings.IndexType == "full" && settings.ConvertToMiniOcr == false {
		// use single quotes to submit the XML in solr post
		out = strings.ReplaceAll(out, "\"", "'")
	}

	return &out, nil
}

// updateXML sets the hOCR page ID and converts unicode to XML-escaped codepoints when require by configuration.
func updateXML(ocr *string, position int, settings model.Configuration) (*string, error) {

	if !settings.EscapeUtf8 && settings.IndexType != "lazy" {
		return ocr, nil
	}

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
		case xml.Comment:
			if err := encoder.EncodeToken(t); err != nil {
				return nil, err
			}
			continue
		case xml.CharData:
			if xmlEncodeWord && len(t) > 0 {
				escaped := []byte(ToXmlCodePoint(string(t)))
				t = escaped
				if err := encoder.EncodeToken(t); err != nil {
					return nil, err
				}
				xmlEncodeWord = false
				continue
			}

		case xml.StartElement:
			if hasClassValue(t, "ocr_page") {
				id := "Page." + strconv.Itoa(position)
				pos := getPosition(t, "id")
				t.Attr[pos].Value = id
				if err := encoder.EncodeToken(t); err != nil {
					return nil, err
				}
				continue
			}

			if hasClassValue(t, "ocrx_word") && settings.EscapeUtf8 && settings.IndexType == "lazy" {
				if err := encoder.EncodeToken(t); err != nil {
					return nil, err
				}
				xmlEncodeWord = true
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
