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
	log.Println(endpoint)
	resp, err := http.Get(endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, err
}

// Fetches the annotation list from dspace
func getAnnotationList(id string) []byte {
	resp, err := http.Get(id)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}
	return body

}

// Fetches a mets file from DSpace
func getMetsXml(url string) (io.Reader, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(body), err
}

// Fetches an alto file from DSpace
func getAltoXml(url string) (string, error) {

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), err
}
