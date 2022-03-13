package index

import "log"

func HandleAction(indexer Indexer, settings *Configuration, uuid *string, logger *log.Logger) error {
	err := indexer.IndexerAction(settings, uuid, logger)
	if err != nil {
		return err
	}
	return nil
}
