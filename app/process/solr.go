package process

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	. "github.com/mspalti/ocrprocessor/err"
	"github.com/mspalti/ocrprocessor/model"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)


// DeleteFromSolr removes all entries from the solr index for a uuid and (if lazy) removes ocr files from disk.
func DeleteFromSolr(settings model.Configuration, uuid string) error {
	manifestUrl := getDSpaceApiEndpoint(settings.DSpaceHost, uuid, "manifest")
	var files []model.Docs
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

// deleteSolrEntries removes all ocr entries for a manifest from the solr index
func deleteSolrEntries(settings model.Configuration, manifestUrl string) error {
	deleteEndPoint := fmt.Sprintf("%s/%s/update?", settings.SolrUrl, settings.SolrCore)
	deleteByManifest := url.QueryEscape("\"" + manifestUrl + "\"")
	deleteBody := "manifest_url:" + deleteByManifest
	solrPostBody := &model.SolrDeletePost{
		Delete: model.Delete{Query: deleteBody},
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

// getFiles returns the indexed ocr file pointers for the manifest (limit 600 files)
func getFiles(settings model.Configuration, manifestUrl string) ([]model.Docs, error) {
	solrUrl := fmt.Sprintf("%s/%s/select?fl=ocr_text&rows=600&q=manifest_url:%s",
		settings.SolrUrl, settings.SolrCore, url.QueryEscape("\""+manifestUrl+"\""))
	payloadBuf := new(bytes.Buffer)
	req, err := http.NewRequest("GET", solrUrl, payloadBuf)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.New("could not query solr for files to delete: " + err.Error())
	}
	defer resp.Body.Close()
	solrResponse := model.SolrResponse{}
	err = json.NewDecoder(resp.Body).Decode(&solrResponse)
	files := solrResponse.Response.Docs
	return files, nil
}

// deleteFiles removes the ocr files on disk
func deleteFiles(files []model.Docs) error {
	for i := 0; i < len(files); i++ {
		file := files[i].OcrText
		file = strings.Replace(file, "{ascii}", "", 1)
		err := os.Remove(file)
		if err != nil {
			return err
		}
	}
	return nil
}

// CheckSolr returns true if the index has entries for the uuid
func CheckSolr(settings model.Configuration, uuid string) (bool, error) {
	manifestUrl := getDSpaceApiEndpoint(settings.DSpaceHost, uuid, "manifest")
	solrUrl := fmt.Sprintf("%s/%s/select?fl=manifest_url&q=manifest_url:%s",
		settings.SolrUrl, settings.SolrCore, url.QueryEscape("\""+manifestUrl+"\""))
	payloadBuf := new(bytes.Buffer)
	req, err := http.NewRequest("GET", solrUrl, payloadBuf)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, errors.New("could not query solr: " + err.Error())
	}
	defer resp.Body.Close()

	solrResponse := model.SolrResponse{}
	err = json.NewDecoder(resp.Body).Decode(&solrResponse)
	if err != nil {
		return false, err
	}
	if solrResponse.Response.NumFound > 0 {
		return true, nil
	}
	return false, nil
}

// PostToSolrLazyLoad adds to solr index and writes alto file to disk. Alto file will be lazy loaded by the solr plugin
func PostToSolrLazyLoad(uuid *string, fileName string, altoFile *string, manifestId string,
	settings model.Configuration, log *log.Logger) error {
	var extension = filepath.Ext(fileName)
	solrId := *uuid + "-" + fileName[0:len(fileName)-len(extension)]
	path := settings.XmlFileLocation + "/" + solrId + ".xml"
	err2 := ioutil.WriteFile(path, []byte(*altoFile), 0644)
	if err2 != nil {
		return errors.New("could not write escaped alto file")
	}
	if settings.EscapeUtf8 {
		path = path + "{ascii}"
	}
	solrPostBody := &model.SolrCreatePost{
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

// PostToSolr add the miniOcr content directly to the solr index. No lazy loading.
func PostToSolr(uuid *string, fileName string, miniOcr *string, manifestId string,
	settings model.Configuration, log *log.Logger) error {
	var extension = filepath.Ext(fileName)
	solrId := *uuid + "-" + fileName[0:len(fileName)-len(extension)]
	solrPayload := &model.SolrCreatePost{
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
