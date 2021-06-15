package index

import (
	"bytes"
	"io"
	"net/http"
)

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
