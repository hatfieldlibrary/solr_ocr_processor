package index

import (
	"bytes"
	"io"
	"log"
	"net/http"
)

func getApiEndpoint(host string, uuid string, iiiftype string) string {
	return host + "/iiif/" + uuid + "/" + iiiftype
}

// Fetches manifest from dspace
func getManifest(host string, uuid string, log *log.Logger) ([]byte, error) {
	endpoint := getApiEndpoint(host, uuid, "manifest")

	resp, err := http.Get(endpoint)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	if resp.StatusCode != 200 {
		errorMessage := UnProcessableEntity{"could not retrieve manifest. Status:  " + resp.Status}
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

// Fetches the annotation list from dspace
func getAnnotationList(id string, log *log.Logger) ([]byte, error) {
	resp, err := http.Get(id)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	if resp.StatusCode != 200 {
		errorMessage := UnProcessableEntity{"could not retrieve annotations. Status:  " + resp.Status}
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

// getMetsXml fetches a mets file from DSpace
func getMetsXml(url string, log *log.Logger) (io.Reader, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	if resp.StatusCode != 200 {
		errorMessage := UnProcessableEntity{"could not retrieve mets xml. Status:  " + resp.Status}
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

// getAltoXml fetches an alto file from DSpace
func getAltoXml(url string, log *log.Logger) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		log.Println(err.Error())
		errorMessage := UnProcessableEntity{"could not retrieve alto xml. Status:  " + resp.Status}
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
