# Alto Indexer

This is an early release for testing.

## Processing DSpace IIIF Records for Search API Requests 
This service pre-processes METS/ALTO files for indexing by the solr-ocrhighlighting Solr plugin. That plugin is maintained by the MDZ Digital Library team: https://github.com/dbmdz/solr-ocrhighlighting.

When indexing a new item, the service takes the DSpace Item ID as an HTTP request parameter. It then retrieves the IIIF `Manifest` and an `AnnotationList` that describes the METS and ALTO files for the DSpace item, including URLs for retrieving the files from DSpace. The METS file provides the structure of the document, including the sequence of ALTO OCR files. This information is used to retrieve and post ALTO files to the Solr plugin for indexing, in page order. The ALTO files are also preprocessed and written to disk as required by the Solr plugin.

This service must run on the same file system as Solr.

## Installation

A download for distribution will be provided soon. In the meantime, a Docker container is available for early testing.

`docker pull mspalti/altoindexer:latest`

To run the container:

`docker run -d --network host -v /host/path/to/configs:/app/configs -v /host/path/to/logs:/app/logs -v /path/escaped/alto/files/:/var/escaped_alto_files mspalti/altoindexer`

## Usage

Currently, only "add" operations are supported. The service will soon support deletion. At this time the only supported method is GET but POST and DELETE methods are planned.  

`http://<host>:3000/413065ef-e242-4d0e-867d-8e2f6486be56/add`



