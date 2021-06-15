package index

func HandleAction(settings Configuration, uuid *string, action *string) error {

	if *action == "add" && len(*uuid) > 0 {
		err := AddItem(settings, *uuid)
		if err != nil {
			return err
		}
	}

	return nil
}
