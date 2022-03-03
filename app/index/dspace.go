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
func getManifest(host string, uuid string) ([]byte, error) {
	endpoint := getApiEndpoint(host, uuid, "manifest")

	resp, err := http.Get(endpoint)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		errorMessage := UnProcessableEntity{"could not retrieve manifest. Status:  " + resp.Status}
		return nil, errorMessage
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, err
}

// Fetches the annotation list from dspace
func getAnnotationList(id string) ([]byte, error) {
	resp, err := http.Get(id)
	if err != nil {
		log.Println(err)
	}
	if resp.StatusCode != 200 {
		errorMessage := UnProcessableEntity{"could not retrieve annotations. Status:  " + resp.Status}
		return nil, errorMessage
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}
	return body, nil

}

// Fetches a mets file from DSpace
func getMetsXml(url string) (io.Reader, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		errorMessage := UnProcessableEntity{"could not retrieve mets xml. Status:  " + resp.Status}
		return nil, errorMessage
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(body), err
}

// Fetches an alto file from DSpace
func getAltoXml(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		errorMessage := UnProcessableEntity{"could not retrieve alto xml. Status:  " + resp.Status}
		return nil, errorMessage
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// return string(body), err
	return body, err
}
