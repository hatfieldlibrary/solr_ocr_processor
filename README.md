# Alto Indexer

This is an early release for testing.

## Processing DSpace IIIF Records for Search API Requests 
This service pre-processes METS/ALTO files for indexing by the solr-ocrhighlighting Solr plugin. That plugin is maintained by the MDZ Digital Library team: https://github.com/dbmdz/solr-ocrhighlighting.

The service takes the DSpace Item ID as an HTTP request parameter. 

It retrieves the IIIF `Manifest` and an `AnnotationList` that the references METS and ALTO files for the DSpace item. 

The METS file provides the sequence of ALTO OCR files. This information is used to retrieve the ALTO files from DSpace,
process the files, and POST them to the 
Solr plugin for indexing. If lazy loading, files are also written to disk. When lazy loading, the Solr service 
MUST be able to access the shared file system.

## Installation

A download for distribution will be provided soon. In the meantime, a Docker container is available for early testing.

`docker pull mspalti/altoindexer:latest`

To run the container:

`docker run -d --network host -v /host/path/to/configs:/app/configs -v /host/path/to/logs:/app/logs -v /path/escaped/alto/files:/var/escaped_alto_files mspalti/altoindexer`

## Usage

Currently, only "add" operations are supported. The service will soon support search and deletion. At this time the only supported method is GET but POST and DELETE methods are planned.  

`http://<host>:3000/413065ef-e242-4d0e-867d-8e2f6486be56/add`



