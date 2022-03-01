package index

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
)

// postToSolrLazyLoad adds to solr index and writes alto file to disk. Alto file will be lazy loaded by the solr plugin
func postToSolrLazyLoad(uuid string, fileName string, altoFile *string, manifestId string,
	settings Configuration) error {

	var extension = filepath.Ext(fileName)
	solrId := uuid + "-" + fileName[0:len(fileName)-len(extension)]
	path := settings.XmlFileLocation + "/" + solrId + "_escaped.xml"

	err2 := ioutil.WriteFile(path, []byte(*altoFile), 0644)
	if err2 != nil {
		return errors.New("could not write escaped alto file")
	}

	solrPostBody := &SolrPost{
		Id:          solrId,
		ManifestUrl: manifestId,
		OcrText:     path}

	payloadBuf := new(bytes.Buffer)
	json.NewEncoder(payloadBuf).Encode(solrPostBody)
	solrUrl := fmt.Sprintf("%s/%s/update/json/docs", settings.SolrUrl, settings.SolrCore)

	req, err := http.NewRequest("POST", solrUrl, payloadBuf)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return errors.New("could not post to solr file")
	}
	defer resp.Body.Close()

	return nil

}

// postToSolr add the miniOcr content directly to the solr index. No lazy loading.
func postToSolr(uuid string, fileName string, miniOcr *string, manifestId string,
	settings Configuration) error {
	var extension = filepath.Ext(fileName)
	solrId := uuid + "-" + fileName[0:len(fileName)-len(extension)]
	solrPayload := &SolrPost{
		Id:          solrId,
		ManifestUrl: manifestId,
		OcrText:     *miniOcr}
	payloadBuf := new(bytes.Buffer)
	enc := json.NewEncoder(payloadBuf)
	enc.SetEscapeHTML(false)
	enc.Encode(solrPayload)
	solrUrl := fmt.Sprintf("%s/%s/update/json/docs", settings.SolrUrl, settings.SolrCore)

	req, err := http.NewRequest("POST", solrUrl, payloadBuf)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return errors.New("could not post to solr file")
	}
	defer resp.Body.Close()
	return nil
}
