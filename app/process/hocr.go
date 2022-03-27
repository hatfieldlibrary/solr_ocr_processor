package process

import (
	. "github.com/mspalti/altoindexer/err"
	"github.com/mspalti/altoindexer/model"
	"log"
)

func (processor HocrProcessor) ProcessOcr(uuid *string, fileName string, alto *string, position int,
	manifestId string, settings model.Configuration, log *log.Logger) error {

	return UnProcessableEntity{CAUSE: "processing for hOCR format is not implemented"}
}
