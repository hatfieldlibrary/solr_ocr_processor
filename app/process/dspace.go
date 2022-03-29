package process

import (
	"bytes"
	. "github.com/mspalti/ocrprocessor/err"
	"io"
	"log"
	"net/http"
)

func getApiEndpoint(host string, uuid string, iiiftype string) string {
	return host + "/iiif/" + uuid + "/" + iiiftype
}

// GetManifest fetches the manifest from dspace
func GetManifest(host string, uuid string, log *log.Logger) ([]byte, error) {
	endpoint := getApiEndpoint(host, uuid, "manifest")

	resp, err := http.Get(endpoint)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	if resp.StatusCode != 200 {
		errorMessage := UnProcessableEntity{CAUSE: "could not retrieve manifest. Status:  " + resp.Status}
		return nil, errorMessage
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	return body, err
}

// GetAnnotationList fetches the annotation list from dspace
func GetAnnotationList(id string, log *log.Logger) ([]byte, error) {
	resp, err := http.Get(id)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	if resp.StatusCode != 200 {
		errorMessage := UnProcessableEntity{CAUSE: "could not retrieve annotations. Status:  " + resp.Status}
		return nil, errorMessage
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	return body, nil

}

// GetMetsXml fetches a mets file from DSpace
func GetMetsXml(url string, log *log.Logger) (io.Reader, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	if resp.StatusCode != 200 {
		errorMessage := UnProcessableEntity{CAUSE: "could not retrieve mets xml. Status:  " + resp.Status}
		return nil, errorMessage
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	return bytes.NewReader(body), err
}

// GetOcrXml fetches an alto file from DSpace
func GetOcrXml(url string, log *log.Logger) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		log.Println(err.Error())
		errorMessage := UnProcessableEntity{CAUSE: "could not retrieve alto xml. Status:  " + resp.Status}
		return "", errorMessage
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err.Error())
		return "", err
	}
	return string(body), err
}
