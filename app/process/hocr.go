package process

import (
	. "github.com/mspalti/ocrprocessor/err"
	"github.com/mspalti/ocrprocessor/model"
	"log"
)

func (processor HocrProcessor) ProcessOcr(uuid *string, fileName string, alto *string, position int,
	manifestId string, settings model.Configuration, log *log.Logger) error {

	return UnProcessableEntity{CAUSE: "processing for hOCR format is not implemented"}
}
