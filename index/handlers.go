package index


func HandleAction(indexer Indexer, settings *Configuration, uuid *string) error {

	err := indexer.IndexerAction(settings, uuid)
	if err != nil {
		return err
	}

	return nil
}


