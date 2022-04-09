package process

import (
	"github.com/mspalti/ocrprocessor/model"
	"log"
)

type OcrProcessor interface {
	// ProcessOcr implements transformations and loading of OCR files.
	ProcessOcr(uuid *string, fileName string, alto *[]byte, position int,
		manifestId string, settings model.Configuration, log *log.Logger) error
}

type AltoProcessor struct{}
type HocrProcessor struct{}
type MiniOcrProcessor struct{}
