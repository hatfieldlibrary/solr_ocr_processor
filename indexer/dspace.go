package main

import (
	"bytes"
	"io"
	"net/http"
)

func getMetsXml(url string) io.Reader {
	resp, err := http.Get(url)
	if err != nil {
		println("oops")
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		println("oops")
	}

	return bytes.NewReader(body)
}

func getAltoXml(url string) string {

	resp, err := http.Get(url)
	if err != nil {
		println("oops")
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return ""
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		println("oops")
	}

	return string(body)
}
