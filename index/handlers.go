package index


func HandleAction(indexers []Indexer, settings *Configuration, uuid *string, action *string) error {

	if *action == "add" && len(*uuid) > 0 {
		// add item interface
		err := indexers[0].IndexerAction(settings, uuid)
		if err != nil {
			return err
		}
	}
	if *action == "delete" && len(*uuid) > 0 {
		// delete item interface
		err := indexers[1].IndexerAction(settings, uuid)
		if err != nil {
			return err
		}
	}
	return nil
}


