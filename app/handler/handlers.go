package handler

import (
	"github.com/mspalti/ocrprocessor/model"
	"log"
)

func HandleAction(indexer Indexer, settings *model.Configuration, uuid *string, logger *log.Logger) error {
	err := indexer.IndexerAction(settings, uuid, logger)
	if err != nil {
		return err
	}
	return nil
}
