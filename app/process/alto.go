package process

import (
	"bytes"
	"encoding/xml"
	"errors"
	"github.com/mspalti/ocrprocessor/model"
	"io"
	"log"
	"regexp"
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

	// There is no need to update when full indexing or no character conversion is requested.
	if settings.IndexType != "lazy" && !settings.EscapeUtf8 {
		out := string(*alto)
		return &out, nil
	}

	var buffer bytes.Buffer
	reader := bytes.NewReader(*alto)
	decoder := xml.NewDecoder(reader)
	encoder := xml.NewEncoder(&buffer)

	var dpiMatcher = regexp.MustCompile(`xdpi:(\d+)`)

	// These control conversion from inch1200 to pixel units
	// This will be attempted for every ALTO file in which the
	// MeasurementUnit is set to 'inch1200'
	checkUnit := false
	convertInchToPixel := false
	convertMM10ToPixel := false
	lookForDpi := false
	dpiValue := -1

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

			str := string(t)
			if checkUnit {
				if str == "inch1200" {
					convertInchToPixel = true
					t = []byte("pixel")
				}
				if str == "mm10" {
					convertMM10ToPixel = true
					t = []byte("pixel")
				}
				checkUnit = false
			}
			if lookForDpi {
				dpi := dpiMatcher.FindSubmatch([]byte(str))
				dpiValue, err = strconv.Atoi(string(dpi[1]))
				if err != nil {
					return nil, err
				}
				lookForDpi = false
			}

		case xml.StartElement:
			if t.Name.Local == "MeasurementUnit" {
				checkUnit = true
			}
			if t.Name.Local == "processingStepSettings" {
				lookForDpi = true
			}
			if t.Name.Local == "Page" {
				id := "Page." + strconv.Itoa(position)
				idPos := getPosition(t, "ID")
				t.Attr[idPos].Value = id
				if convertInchToPixel {
					err := inchToPixel(&t, dpiValue, settings)
					if err != nil {
						return nil, err
					}
				}
				if convertMM10ToPixel {
					err := mmToPixel(&t)
					if err != nil {
						return nil, err
					}
				}
				if err := encoder.EncodeToken(t); err != nil {
					return nil, err
				}
				continue
			}

			if t.Name.Local == "String" {
				modified := false
				if settings.EscapeUtf8 && settings.IndexType == "lazy" {
					pos := getPosition(t, "CONTENT")
					t.Attr[pos].Value = ToXmlCodePoint(t.Attr[pos].Value)
					modified = true
				}
				if convertMM10ToPixel || convertInchToPixel {
					if convertInchToPixel {
						err := inchToPixel(&t, dpiValue, settings)
						if err != nil {
							return nil, err
						}
					}
					if convertMM10ToPixel {
						err := mmToPixel(&t)
						if err != nil {
							return nil, err
						}
					}
					modified = true
				}
				// If String token values were modified then encode now and continue.
				if modified {
					if err := encoder.EncodeToken(t); err != nil {
						return nil, err
					}
					continue
				}
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

func inchToPixel(t *xml.StartElement, dpiValue int, settings model.Configuration) error {

	h := getPosition(*t, "HEIGHT")
	w := getPosition(*t, "WIDTH")
	harr := strings.Split(t.Attr[h].Value, ".")
	warr := strings.Split(t.Attr[w].Value, ".")
	height, err := strconv.Atoi(harr[0])
	if err != nil {
		return err
	}
	width, err := strconv.Atoi(warr[0])
	if err != nil {
		return err
	}
	t.Attr[h].Value = convertInchDimToPixel(height, dpiValue, settings)
	t.Attr[w].Value = convertInchDimToPixel(width, dpiValue, settings)

	if t.Name.Local == "String" {
		hp := getPosition(*t, "HPOS")
		vp := getPosition(*t, "VPOS")
		hposarr := strings.Split(t.Attr[hp].Value, ".")
		vposarr := strings.Split(t.Attr[vp].Value, ".")
		hpos, err := strconv.Atoi(hposarr[0])
		if err != nil {
			return err
		}
		vpos, err := strconv.Atoi(vposarr[0])
		if err != nil {
			return err
		}
		t.Attr[hp].Value = convertInchDimToPixel(hpos, dpiValue, settings)
		t.Attr[vp].Value = convertInchDimToPixel(vpos, dpiValue, settings)
	}

	return nil
}

func convertInchDimToPixel(input int, dpi int, settings model.Configuration) string {
	if dpi == -1 {
		dpi = settings.InputImageResolution
	}
	dim := (input * dpi) / 1200
	return strconv.Itoa(dim)
}

// mmToPixel updates mm10 unit values to pixels
func mmToPixel(t *xml.StartElement) error {
	h := getPosition(*t, "HEIGHT")
	w := getPosition(*t, "WIDTH")
	hp := getPosition(*t, "HPOS")
	vp := getPosition(*t, "VPOS")
	htmm, err := strconv.Atoi(t.Attr[h].Value)
	if err != nil {
		return err
	}
	wdmm, err := strconv.Atoi(t.Attr[w].Value)
	if err != nil {
		return err
	}
	var hposmm *int
	var vposmm *int
	if hp >= 0 {
		v, err := strconv.Atoi(t.Attr[hp].Value)
		hposmm = &v
		if err != nil {
			return err
		}
	}
	if vp >= 0 {
		v, err := strconv.Atoi(t.Attr[vp].Value)
		vposmm = &v
		if err != nil {
			return err
		}
	}
	height := 3.7795275591 * float64(htmm)
	width := 3.7795275591 * float64(wdmm)
	t.Attr[h].Value = strconv.Itoa(int(height))
	t.Attr[w].Value = strconv.Itoa(int(width))

	if hposmm != nil {
		hpos := 3.7795275591 * float64(*hposmm)
		t.Attr[hp].Value = strconv.Itoa(int(hpos))
	}
	if vposmm != nil {
		vpos := 3.7795275591 * float64(*vposmm)
		t.Attr[vp].Value = strconv.Itoa(int(vpos))
	}

	return nil
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
				h := getPosition(t, "HEIGHT")
				w := getPosition(t, "WIDTH")
				height := t.Attr[h].Value
				width := t.Attr[w].Value
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
				c := getPosition(t, "CONTENT")
				h := getPosition(t, "HEIGHT")
				w := getPosition(t, "WIDTH")
				hp := getPosition(t, "HPOS")
				vp := getPosition(t, "VPOS")

				content := t.Attr[c]
				height := t.Attr[h]
				width := t.Attr[w]
				hpos := t.Attr[hp]
				vpos := t.Attr[vp]

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
