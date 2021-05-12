package index

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
)

func postToSolr(uuid string, fileName string, altoFile string, manifestId string, identifier string, settings Configuration) {

	var extension = filepath.Ext(fileName)
	solrId := uuid + "-" + fileName[0:len(fileName)-len(extension)]
	path := settings.XmlFileLocation + "/" + solrId + "_escaped.xml"

	err2 := ioutil.WriteFile(path, []byte(altoFile), 0644)
	if err2 != nil {
		panic(err2)
	}

	solrPostBody := &SolrPost{
		Id:          solrId,
		ManifestUrl: manifestId,
		OcrText:     path}

	payloadBuf := new(bytes.Buffer)
	json.NewEncoder(payloadBuf).Encode(solrPostBody)
	solrUrl := fmt.Sprintf("%s/%s/update/json/docs", settings.SolrUrl, settings.SolrCore)

	// fmt.Println(payloadBuf)

	req, err := http.NewRequest("POST", solrUrl, payloadBuf)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)


}
