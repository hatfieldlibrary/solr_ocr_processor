package index

import "testing"

type SpyFakeAddItem struct {
	settings *Configuration
	uuid *string
	indexerActionWasCalled bool
}

type SpyFakeDeleteItem struct {
	settings *Configuration
	uuid *string
	indexerActionWasCalled bool
}

func (f *SpyFakeAddItem) IndexerAction(settings *Configuration, uuid *string) error {
	f.indexerActionWasCalled = true
	f.settings = settings
	f.uuid = uuid
	return nil
}

func (f *SpyFakeDeleteItem) IndexerAction(settings *Configuration, uuid *string) error {
	f.indexerActionWasCalled = true
	f.settings = settings
	f.uuid = uuid
	return nil
}

func TestHandleAction(t *testing.T) {
	uuid := "1243"
	configuration := Configuration{
		DSpaceHost:      "",
		Collections:     nil,
		SolrUrl:         "",
		SolrCore:        "",
		XmlFileLocation: "",
		HttpPort:        "",
		LogDir:          "",
	}

	spy := &SpyFakeAddItem{settings: &configuration, uuid: &uuid}
	spydel := &SpyFakeDeleteItem{settings: &configuration, uuid: &uuid}
	indexers := []Indexer{spy, spydel}

	// test add
	axn := "add"
	err := HandleAction(indexers, &configuration, &uuid, &axn)
	if err != nil {
		print(err)
	}
	if !spy.indexerActionWasCalled {
		t.Errorf("expected call to indexer using add")
	}

	// test delete
	axn = "delete"
	err = HandleAction(indexers, &configuration, &uuid, &axn)
	if err != nil {
		print(err)
	}
	if !spydel.indexerActionWasCalled {
		t.Errorf("expected call to indexer using add")
	}


}
