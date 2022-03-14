package index

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)


// deleteFromSolr removes all entries from the solr index for a uuid and (if lazy) removes ocr files from disk.
func deleteFromSolr(settings Configuration, uuid string) error {
	manifestUrl := getApiEndpoint(settings.DSpaceHost, uuid, "manifest")

	var files []Docs
	var fileError error
	if settings.IndexType == "lazy" {
		files, fileError = getFiles(settings, manifestUrl)
		if fileError != nil {
			return fileError
		}
	}

	err := deleteSolrEntries(settings, manifestUrl)
	if err != nil {
		return err
	}

	if settings.IndexType == "lazy" {
		err = deleteFiles(files)
		if err != nil {
			return err
		}
	}

	return nil
}

// deleteSolrEntries removes all entries for a manifest from the solr index
func deleteSolrEntries(settings Configuration, manifestUrl string) error {
	deleteEndPoint := fmt.Sprintf("%s/%s/update?", settings.SolrUrl, settings.SolrCore)
	deleteByManifest := url.QueryEscape("\""+manifestUrl+"\"")
	deleteBody := "manifest_url:" + deleteByManifest
	solrPostBody := &SolrDeletePost{
		Delete: Delete{Query: deleteBody},
	}
	payloadBuf := new(bytes.Buffer)
	json.NewEncoder(payloadBuf).Encode(solrPostBody)
	req, err := http.NewRequest("POST", deleteEndPoint, payloadBuf)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return errors.New("could not delete solr file: " + err.Error())
	}
	defer resp.Body.Close()

	return nil
}

// getFiles returns the ocr files for this manifest (limit 600 files)
func getFiles(settings Configuration, manifestUrl string) ([]Docs, error) {
	solrUrl := fmt.Sprintf("%s/%s/select?fl=ocr_text&rows=600&q=manifest_url:%s",
		settings.SolrUrl, settings.SolrCore, url.QueryEscape("\"" + manifestUrl + "\""))
	payloadBuf := new(bytes.Buffer)
	req, err := http.NewRequest("GET", solrUrl, payloadBuf)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.New("could not query solr for files to delete: " + err.Error())
	}
	defer resp.Body.Close()
	solrResponse := SolrResponse{}
	err = json.NewDecoder(resp.Body).Decode(&solrResponse)
	files := solrResponse.Response.Docs
	return files, nil
}

// deleteFiles removes the ocr files for a manifest
func deleteFiles(files []Docs) error {
	for i := 0; i < len(files); i++ {
		file := files[i].OcrText
		err := os.Remove(file)
		if err != nil {
			return err
		}
	}
	return nil
}

// checkSolr returns true if the index has entries for the uuid
func checkSolr (settings Configuration, uuid string) (bool, error) {
	manifestUrl := getApiEndpoint(settings.DSpaceHost, uuid, "manifest")

	solrUrl := fmt.Sprintf("%s/%s/select?fl=manifest_url&q=manifest_url:%s",
		settings.SolrUrl, settings.SolrCore, url.QueryEscape("\"" + manifestUrl + "\""))
	payloadBuf := new(bytes.Buffer)
	req, err := http.NewRequest("GET", solrUrl, payloadBuf)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, errors.New("could not query solr: " + err.Error())
	}
	defer resp.Body.Close()

	solrResponse := SolrResponse{}
	err = json.NewDecoder(resp.Body).Decode(&solrResponse)
	if err != nil {
		return false, err
	}
	if solrResponse.Response.NumFound > 0 {
		return true, nil
	}
	return false, nil
}

// postToSolrLazyLoad adds to solr index and writes alto file to disk. Alto file will be lazy loaded by the solr plugin
func postToSolrLazyLoad(uuid string, fileName string, altoFile *string, manifestId string,
	settings Configuration, log *log.Logger) error {

	var extension = filepath.Ext(fileName)
	solrId := uuid + "-" + fileName[0:len(fileName)-len(extension)]
	path := settings.XmlFileLocation + "/" + solrId + ".xml"
	err2 := ioutil.WriteFile(path, []byte(*altoFile), 0644)
	if err2 != nil {
		return errors.New("could not write escaped alto file")
	}
	if settings.EscapeUtf8 {
		path = path + "{ascii}"
	}
	solrPostBody := &SolrCreatePost{
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
		log.Println(err.Error())
		return UnProcessableEntity{CAUSE: "Solr update problem. See log."}
	}
	defer resp.Body.Close()

	return nil

}

// postToSolr add the miniOcr content directly to the solr index. No lazy loading.
func postToSolr(uuid string, fileName string, miniOcr *string, manifestId string,
	settings Configuration, log *log.Logger ) error {
	var extension = filepath.Ext(fileName)
	solrId := uuid + "-" + fileName[0:len(fileName)-len(extension)]
	solrPayload := &SolrCreatePost{
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
		log.Println(err.Error())
		return UnProcessableEntity{CAUSE: "Solr update problem. See log."}
	}
	defer resp.Body.Close()
	return nil
}
