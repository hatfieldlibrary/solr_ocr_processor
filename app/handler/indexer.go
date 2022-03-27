package handler

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	. "github.com/mspalti/altoindexer/err"
	"github.com/mspalti/altoindexer/model"
	"github.com/mspalti/altoindexer/process"
	"io"
	"log"
)

type Indexer interface {
	IndexerAction(settings *model.Configuration, uuid *string, log *log.Logger) error
}

type GetItem struct{}

type AddItem struct{}

type DeleteItem struct{}

// IndexerAction checks whether item is already in the Solr index.
func (axn GetItem) IndexerAction(settings *model.Configuration, uuid *string, log *log.Logger) error {
	exists, err := process.CheckSolr(*settings, *uuid)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	if !exists {
		// If the item is not in index return 404 error code.
		return NotFound{ID: *uuid}
	}
	return nil
}

// IndexerAction adds item to Solr index and writes to file system if lazy loading is used
func (axn AddItem) IndexerAction(settings *model.Configuration, uuid *string, log *log.Logger) error {
	manifestJson, err := process.GetManifest(settings.DSpaceHost, *uuid, log)
	if err != nil {
		return err
	}
	manifest, err := unMarshallManifest(manifestJson)
	if err != nil {
		return err
	}
	// retrieve the iiif seeAlso annotation list from DSpace
	annotationListJson, err := process.GetAnnotationList(manifest.SeeAlso.Id, log)
	if err != nil {
		return err
	}
	annotations, err := unMarshallAnnotationList(annotationListJson)
	if err != nil {
		return err
	}
	// for each Resource, create a map with the file name as the key and the iiif identifier as the value
	annotationsMap := createAnnotationMap(annotations.Resources)
	if len(annotationsMap) == 0 {
		err := UnProcessableEntity{CAUSE: "no annotations exist for this item, nothing to process"}
		return err
	}
	// Create ordered list of file names using either the METS file (named mets.xml) or the DSpace bundle's bitstream
	// order. The processing order defines page identifiers that must match canvas identifiers in the
	// IIIF manifest. If these do not align, search results and word highlighting will be incorrect. The METS
	// file is a good way to assure correct order. Without it, you must guarantee that OCR bitstreams in the DSpace
	// OtherContent bundle (used for seeAlso annotations) are ordered correctly.
	var ocrFiles []string
	if metsReader, err := getMetsFileReader(annotationsMap["mets.xml"], log); err == nil {
		ocrFiles = getMetsOcrFileNames(metsReader)
	} else {
		ocrFiles = getOcrFilesFromAnnotationList(annotations.Resources)
	}
	var format process.Format
	// traverse though ordered list of file names
	for i := 0; i < len(ocrFiles); i++ {
		var ocr string
		if len(ocrFiles[i]) > 0 {
			// fetch the OCR file from DSpace
			ocr, err = process.GetOcrXml(annotationsMap[ocrFiles[i]], log)
			if err != nil {
				return err
			}
			// get the OCR file format
			format = process.GetOcrFormat(ocr[0:140])
		}
		if len(ocr) != 0 {
			log.Println("processing OCR format: " + format.String())
			var processor process.OcrProcessor
			switch format {
			case process.AltoFormat:
				processor = process.AltoProcessor{}
			case process.HocrFormat:
				processor = process.HocrProcessor{}
			case process.MiniocrFormat:
				processor = process.MiniOcrProcessor{}
			case process.UnknownFormat:
				log.Printf("ignoring %s file format", format.String())
			}
			if processor != nil {
				err := processor.ProcessOcr(uuid, ocrFiles[i], &ocr, i, manifest.Id, *settings, log)
				if err != nil {
					log.Printf("OCR processing failure for %s: %s", ocrFiles[i], err.Error())
					return err
				}
			}

		}
	}
	return nil
}

// IndexerAction deletes all records for manifest from the Solr index and deletes files if lazy loading is used.
func (axn DeleteItem) IndexerAction(settings *model.Configuration, uuid *string, log *log.Logger) error {
	err := process.DeleteFromSolr(*settings, *uuid)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	log.Println("Deleted from index: " + *uuid)
	return nil
}

// getMetsFileReader returns a byte reader for the METS file found in DSpace or an error if the file is not found
func getMetsFileReader(identifier string, log *log.Logger) (io.Reader, error) {
	if len(identifier) == 0 {
		return nil, errors.New("an iiif identifier was not found for the mets.xml file")
	}
	metsReader, err := process.GetMetsXml(identifier, log)
	if err != nil {
		return nil, err
	}
	return metsReader, nil
}

// getOcrFilesFromAnnotationList creates the array of file names from the ResourceAnnotation list
func getOcrFilesFromAnnotationList(annotations []model.ResourceAnnotation) []string {
	arr := make([]string, 0)
	for i := 0; i < len(annotations); i++ {
		arr = append(arr, annotations[i].Resource.Label)
	}
	return arr
}

// getMetsOcrFileNames creates the array of file names from the METS file.
func getMetsOcrFileNames(metsReader io.Reader) []string {
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

// createAnnotationMap creates a map with the IIIF label as key and IIIF resource id as the value
func createAnnotationMap(annotations []model.ResourceAnnotation) map[string]string {
	annotationMap := make(map[string]string)
	for i := 0; i < len(annotations); i++ {
		value := annotations[i].Resource.Id
		key := annotations[i].Resource.Label
		annotationMap[key] = value
	}
	return annotationMap
}

func unMarshallManifest(bytes []byte) (model.Manifest, error) {
	var manifest model.Manifest
	if err := json.Unmarshal(bytes, &manifest); err != nil {
		errorMessage := errors.New("could not unmarshal manifest: " + err.Error())
		return manifest, errorMessage
	}
	return manifest, nil
}

func unMarshallAnnotationList(bytes []byte) (model.ResourceAnnotationList, error) {
	var annotations model.ResourceAnnotationList
	if err := json.Unmarshal(bytes, &annotations); err != nil {
		errorMessage := errors.New("could not unmarshal annotationlist: " + err.Error())
		return annotations, errorMessage
	}
	return annotations, nil
}
