package main

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
)

func AddToIndex(settings Configuration, uuid string) {
	manifest := unMarshallManifest(getManifest(settings.DSpaceHost, uuid))
	annotations := unMarshallAnnotationList(getAnnotationList(manifest.SeeAlso.Id))
	annotationsMap := createAnnotationHash(annotations.Resources)
	altoFiles := getAltoFiles(annotationsMap)
	indexFiles(uuid, annotationsMap, altoFiles, manifest.Id, settings)
}

func getAltoFiles(annotationsMap map[string]string) []string {
	metsReader := getMetsXml(annotationsMap["mets.xml"])
	altoFiles := getOcrFileNames(metsReader)
	return altoFiles
}


func getOcrFileNames(metsReader io.Reader) []string {
	var fileNames = make([]string, 50)
	parser := xml.NewDecoder(metsReader)
	ocrFileElement := false
	altoCounter := 0
	for {
		token, err := parser.Token()
		if err != nil {
			break
		}
		switch t := token.(type) {
		case xml.StartElement:
			element := xml.StartElement(t)
			name := element.Name.Local
			if name == "file" {
				for i := 0; i < len(element.Attr); i++ {
					if element.Attr[i].Value == "ocr" {
						ocrFileElement = true
					}
				}
			}
			if name == "FLocat" && ocrFileElement == true {
				for i := 0; i < len(element.Attr); i++ {
					if element.Attr[i].Name.Local == "href" {
						// Allocate more capacity.
						if altoCounter == cap(fileNames)  {
							newFileNames := make([]string, 2*cap(fileNames))
							copy(newFileNames, fileNames)
							fileNames = newFileNames
						}
						fileNames[altoCounter] = element.Attr[i].Value
						altoCounter++
					}
				}
			}

		case xml.EndElement:
			element := xml.EndElement(t)
			name := element.Name.Local
			if name == "file" {
				ocrFileElement = false
			}
		}
	}

	return fileNames

}

func createAnnotationHash(annotations []ResourceAnnotation) map[string]string {

	annotationMap := make(map[string]string)
	for i := 0; i < len(annotations); i++ {
		value := annotations[i].Resource.Id
		key := annotations[i].Resource.Label
		annotationMap[key] = value
	}
	return annotationMap
}

func unMarshallManifest(bytes []byte) Manifest {
	var manifest Manifest
	if err := json.Unmarshal(bytes, &manifest); err != nil {
		panic(err)
	}
	return manifest
}

func unMarshallAnnotationList(bytes []byte) ResourceAnnotationList {
	var annotations ResourceAnnotationList
	if err := json.Unmarshal(bytes, &annotations); err != nil {
		panic(err)
	}
	return annotations
}

func getManifest(host string, uuid string) []byte {
	endpoint := getApiEndpoint(host, uuid, "manifest")
	resp, err := http.Get(endpoint)
	if err != nil {
		println(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		println(err)
	}
	return body
}

func getAnnotationList(id string) []byte {
	resp, err := http.Get(id)
	if err != nil {
		println("oops")
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		println(err)
	}
	return body

}

func getApiEndpoint(host string, uuid string, iiiftype string) string {
	return host + "/iiif/" + uuid + "/" + iiiftype
}
