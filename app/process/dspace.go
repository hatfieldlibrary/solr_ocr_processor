package process

import (
	. "github.com/mspalti/ocrprocessor/err"
	"io"
	"log"
	"net/http"
)

// GetManifest fetches the manifest from DSpace
func GetManifest(host string, uuid string, log *log.Logger) ([]byte, error) {
	endpoint := getDSpaceApiEndpoint(host, uuid, "manifest")
	resp, err := http.Get(endpoint)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	if resp.StatusCode != 200 {
		errorMessage := UnProcessableEntity{CAUSE: "Could not retrieve manifest. Status:  " + resp.Status}
		return nil, errorMessage
	}
	return responseReader(resp.Body)
}

// GetAnnotationList fetches the annotation list from DSpace
func GetAnnotationList(id string, log *log.Logger) ([]byte, error) {
	resp, err := http.Get(id)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	if resp.StatusCode != 200 {
		errorMessage := UnProcessableEntity{CAUSE: "Could not retrieve annotations. Status:  " + resp.Status}
		return nil, errorMessage
	}
	return responseReader(resp.Body)
}

// GetMetsXml fetches a mets file from DSpace
func GetMetsXml(url string, log *log.Logger) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	if resp.StatusCode != 200 {
		errorMessage := UnProcessableEntity{CAUSE: "Could not retrieve mets xml. Status:  " + resp.Status}
		return nil, errorMessage
	}
	return responseReader(resp.Body)
}

// GetOcrXml fetches an alto file from DSpace
func GetOcrXml(url string, log *log.Logger) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		log.Println(err.Error())
		errorMessage := UnProcessableEntity{CAUSE: "Could not retrieve OCR file. Status:  " + resp.Status}
		return nil, errorMessage
	}
	return responseReader(resp.Body)
}

func responseReader(reader io.ReadCloser) ([]byte, error) {
	defer reader.Close()
	body, err := io.ReadAll(reader)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	return body, nil
}
