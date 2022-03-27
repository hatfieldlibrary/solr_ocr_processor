package handler

import (
	"github.com/mspalti/altoindexer/model"
	"log"
	"testing"
)

type SpyFakeAddItem struct {
	settings               *model.Configuration
	uuid                   *string
	log                    *log.Logger
	indexerActionWasCalled bool
}

type SpyFakeDeleteItem struct {
	settings               *model.Configuration
	uuid                   *string
	log                    *log.Logger
	indexerActionWasCalled bool
}

func (f *SpyFakeAddItem) IndexerAction(settings *model.Configuration, uuid *string, log *log.Logger) error {
	f.indexerActionWasCalled = true
	f.settings = settings
	f.uuid = uuid
	f.log = log
	return nil
}

func (f *SpyFakeDeleteItem) IndexerAction(settings *model.Configuration, uuid *string, log *log.Logger) error {
	f.indexerActionWasCalled = true
	f.settings = settings
	f.uuid = uuid
	f.log = log
	return nil
}

func TestHandleAction(t *testing.T) {
	uuid := "1243"
	configuration := model.Configuration{
		DSpaceHost:      "",
		Collections:     nil,
		SolrUrl:         "",
		SolrCore:        "",
		XmlFileLocation: "",
		HttpPort:        "",
	}

	// test add
	spy := &SpyFakeAddItem{settings: &configuration, uuid: &uuid, log: log.Default()}
	err := HandleAction(spy, &configuration, &uuid, log.Default())
	if err != nil {
		print(err)
	}
	if !spy.indexerActionWasCalled {
		t.Errorf("expected call to indexer using add")
	}

	// test delete
	spydel := &SpyFakeDeleteItem{settings: &configuration, uuid: &uuid, log: log.Default()}
	err = HandleAction(spydel, &configuration, &uuid, log.Default())
	if err != nil {
		print(err)
	}
	if !spydel.indexerActionWasCalled {
		t.Errorf("expected call to indexer using add")
	}

}
