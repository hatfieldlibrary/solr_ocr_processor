package handler

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	. "github.com/mspalti/ocrprocessor/err"
	"github.com/mspalti/ocrprocessor/model"
	"github.com/mspalti/ocrprocessor/process"
	"io"
	"log"
)

type Indexer interface {
	IndexerAction(settings *model.Configuration, uuid *string, log *log.Logger) error
}

type GetItem struct{}

type AddItem struct{}

type DeleteItem struct{}

// IndexerAction implements the handler interface for GetItem. It is used to test whether OCR files for the
// DSpace Item UUID are already in the Solr index.
func (axn GetItem) IndexerAction(settings *model.Configuration, uuid *string, log *log.Logger) error {
	exists, err := process.CheckSolr(*settings, *uuid)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	if !exists {
		// if the item is not in index return 404 error code.
		return NotFound{ID: *uuid}
	}
	if settings.VerboseLogging {
		log.Printf("This DSpace Item is already in the Solr index: %s", *uuid)
	}
	return nil
}

// IndexerAction implements the handler interface for AddItem. It processes OCR files for a given DSpace
// Item UUID and writes files to disk if lazy loading is requested via configuration. Note that this
// implementation relies on the DSpace IIIF integration to retrieve OCR files for processing.
func (axn AddItem) IndexerAction(settings *model.Configuration, uuid *string, log *log.Logger) error {
	log.Printf("Processing OCR files for DSpace Item: %s", *uuid)
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
	// Processing order determines page identifiers for Solr index entries. These must match Canvas identifiers
	// in the IIIF manifest. If the identifiers do not align then search results and word highlighting will be
	// incorrect. The order of OCR files in the DSpace OtherContent Bundle must therefore match the order of your
	// page images before you attempt Solr indexing.
	//
	// For METS/ALTO projects you can alternately use the order of OCR files found in the METS file. To use the
	// METS file add it to the OtherContent bundle and name the file "mets.xml". If that file is found by
	// the processor the order of processing will be the METS ordering. This can be helpful when the OtherContent
	// Bundle order is inaccurate. Note that the OCR file names in METS and the OtherContent Bundle must be
	// identical when using this approach.
	var ocrFiles []string
	usingMets := false
	if metsReader, err := getMetsFileReader(annotationsMap["mets.xml"], log); err == nil {
		ocrFiles = getMetsOcrFileNames(metsReader)
		usingMets = true
	} else {
		ocrFiles = getOcrFilesFromAnnotationList(annotations.Resources)
	}
	if settings.VerboseLogging {
		if usingMets {
			log.Println("Using the METS file for the processing order.")
			fileCount := len(ocrFiles)
			log.Printf("Processing %d files for the Item %s", fileCount, *uuid)
		}
	}
	var format process.Format
	// initialize the ocr file counter
	var ocrFilePosition = 0
	// traverse though ordered list of file names
	for i := 0; i < len(ocrFiles); i++ {
		var ocr []byte
		if len(ocrFiles[i]) > 0 {
			// fetch the file from DSpace
			ocr, err = process.GetOcrXml(annotationsMap[ocrFiles[i]], log)
			if err != nil {
				log.Printf("Failed to retrieve OCR file from DSpace: %s", annotationsMap[ocrFiles[i]])
				if usingMets {
					log.Println("Check to be sure that the OCR file names in the Bundle match the " +
						"values in your METS file.")
				}
				return err
			}
			ocrString := string(ocr)
			var chunk string
			if len(ocr) > 1200 {
				chunk = ocrString[0:1200]
			} else {
				chunk = ocrString
			}
			// detect the OCR file format based on the 1200 character sample
			format = process.GetOcrFormat(chunk)
		}
		if len(ocr) != 0 {
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
				if settings.VerboseLogging {
					log.Printf("Attempting to process an OCR file in the %s format.", format.String())
				}
				err := processor.ProcessOcr(uuid, ocrFiles[i], &ocr, ocrFilePosition, manifest.Id, *settings, log)
				if err != nil {
					log.Printf("OCR processing failure for %s: %s", ocrFiles[i], err.Error())
					return err
				}
				ocrFilePosition++
			}

		}
	}
	log.Printf("Completed processing item %s with %d OCR files added to the Solr index", *uuid, ocrFilePosition)
	return nil
}

// IndexerAction implements the handler interface for DeleteItem. It deletes all OCR files
// from the Solr index for a given DSpace Item UUID and removes files from disk if lazy loading is used.
func (axn DeleteItem) IndexerAction(settings *model.Configuration, uuid *string, log *log.Logger) error {
	log.Printf("Deleting OCR files for DSpace Item: %s", *uuid)
	err := process.DeleteFromSolr(*settings, *uuid)
	if err != nil {
		log.Printf("Error deleting OCR files from index for the item: %s", err.Error())
		return err
	}
	return nil
}

// getMetsFileReader returns a byte reader for the METS file found in DSpace or an error if the file is not found
func getMetsFileReader(identifier string, log *log.Logger) (io.Reader, error) {
	if len(identifier) == 0 {
		return nil, errors.New("DSpace Bitstream identifier not found for the mets.xml file")
	}
	metsResponse, err := process.GetMetsXml(identifier, log)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(metsResponse), nil
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
