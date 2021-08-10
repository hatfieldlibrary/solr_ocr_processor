package index

import "log"

func HandleAction(indexer Indexer, settings *Configuration, uuid *string) error {
	log.Println(uuid)
	err := indexer.IndexerAction(settings, uuid)
	if err != nil {
		return err
	}

	return nil
}


