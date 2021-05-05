package internal

import (
	"io"
	"net/http"
)

func AddToIndex(host string, uuid string) {
	manifest := getManifest(host, uuid)
	println(manifest)
}

func getManifest(host string, uuid string) string {
	endpoint := getApiEndpoint(host, uuid, "manifest")
	resp, err := http.Get(endpoint)
	if err != nil {
		println("oops")
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		println("bad body")
	}
	manifest := string(body)
	return manifest
}

func getApiEndpoint(host string, uuid string, iiiftype string) string {
	return host + "/iiif/" + uuid + "/" + iiiftype
}
