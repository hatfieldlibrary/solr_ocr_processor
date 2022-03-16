package index

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"log"
)

type Indexer interface {
	IndexerAction(settings *Configuration, uuid *string, log *log.Logger) error
}

type GetItem struct{}

type AddItem struct{}

type DeleteItem struct{}

func (axn GetItem) IndexerAction(settings *Configuration, uuid *string, log *log.Logger) error {
	exists, err := checkSolr(*settings, *uuid)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	if !exists {
		return NotFound{ID: *uuid}
	}
	return nil
}

func (axn AddItem) IndexerAction(settings *Configuration, uuid *string, log *log.Logger) error {
	log.Println("Indexing item: " + *uuid)
	manifestJson, err := getManifest(settings.DSpaceHost, *uuid, log)
	if err != nil {
		return err
	}
	manifest, err := unMarshallManifest(manifestJson)
	if err != nil {
		return err
	}
	annotationListJson, err := getAnnotationList(manifest.SeeAlso.Id, log)
	if err != nil {
		return err
	}
	annotations, err := unMarshallAnnotationList(annotationListJson)
	if err != nil {
		return err
	}
	annotationsMap := createAnnotationMap(annotations.Resources)
	if len(annotationsMap) == 0 {
		errorMessage := UnProcessableEntity{"no annotations exist for this item, nothing to process"}
		return errorMessage
	}
	altoFiles, err := getAltoFiles(annotationsMap, log)
	if err != nil {
		return err
	}
	if settings.FileFormat == "alto" {
		err = processAlto(*uuid, annotationsMap, altoFiles, manifest.Id, *settings, log)
		if err != nil {
			return err
		}
	} else {
		var err = processMiniOcr(*uuid, annotationsMap, altoFiles, manifest.Id, *settings, log)
		if err != nil {
			return err
		}
	}
	return nil
}

func (axn DeleteItem) IndexerAction(settings *Configuration, uuid *string, log *log.Logger) error {
	err := deleteFromSolr(*settings, *uuid)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	return nil
}

// Gets alto file names from mets file.
func getAltoFiles(annotationsMap map[string]string, log *log.Logger) ([]string, error) {
	metsReader, err := getMetsXml(annotationsMap["mets.xml"], log)
	if err != nil {
		return nil, err
	}
	altoFiles := getOcrFileNames(metsReader)
	return altoFiles, nil
}

// getOcrFileNames collects and returns alto file names from the provided mets file reader.
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

			name := t.Name.Local
			if name == "file" {
				for i := 0; i < len(t.Attr); i++ {
					if t.Attr[i].Value == "ocr" {
						ocrFileElement = true
					}
				}
			}
			if name == "FLocat" && ocrFileElement == true {
				for i := 0; i < len(t.Attr); i++ {
					if t.Attr[i].Name.Local == "href" {
						// Allocate more capacity.
						if altoCounter == cap(fileNames) {
							newFileNames := make([]string, 2*cap(fileNames))
							copy(newFileNames, fileNames)
							fileNames = newFileNames
						}
						fileNames[altoCounter] = t.Attr[i].Value
						altoCounter++
					}
				}
			}

		case xml.EndElement:
			name := t.Name.Local
			if name == "file" {
				ocrFileElement = false
			}
		}
	}

	return fileNames

}

// createAnnotationMap creates a map with the label (key) and resource id (value)
func createAnnotationMap(annotations []ResourceAnnotation) map[string]string {
	annotationMap := make(map[string]string)
	for i := 0; i < len(annotations); i++ {
		value := annotations[i].Resource.Id
		key := annotations[i].Resource.Label
		annotationMap[key] = value
	}
	return annotationMap
}

func unMarshallManifest(bytes []byte) (Manifest, error) {
	var manifest Manifest
	if err := json.Unmarshal(bytes, &manifest); err != nil {
		errorMessage := errors.New("could not unmarshal manifest: " + err.Error())
		return manifest, errorMessage
	}
	return manifest, nil
}

func unMarshallAnnotationList(bytes []byte) (ResourceAnnotationList, error) {
	var annotations ResourceAnnotationList
	if err := json.Unmarshal(bytes, &annotations); err != nil {
		errorMessage := errors.New("could not unmarshal annotationlist: " + err.Error())
		return annotations, errorMessage
	}
	return annotations, nil
}
