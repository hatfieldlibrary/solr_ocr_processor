
## Processing Service for the DSpace IIIF Search API
This service pre-processes OCR files for indexing by the `solr-ocrhighlighting` Solr plugin. OCR files are 
retrieved from DSpace using the DSpace IIIF integration. 

**DSpace**: https://wiki.lyrasis.org/display/DSDOC7x

**solr-ocrhighlighting plugin**: https://github.com/dbmdz/solr-ocrhighlighting. 

#### Supports
* GET, POST, and DELETE methods
* Adding `MiniOcr`, `hOCR` or `ALTO` files to the Solr index with "full" or "lazy" indexing (and optional XML-encoding of Unicode characters).
* Conversion of `hOCR` and `ALTO` files to `MiniOcr`.
* Checks for whether OCR files for a DSpace Item have already been indexed.
* Removal of OCR files from the index, and from the file system if "lazy" indexing was used.

#### Configuration Options
* **http_port**: listen port of service
* **ip_whitelist**: IPs that are allowed access
* **dspace_host**: Base URL of the DSpace service
* **solr_url**: Base URL of the Solr service
* **solr_core**: Solr core ("word_highlighting")
* **miniocr_conversion**: Convert OCR to MiniOcr format
* **index_type**: Full or lazy
* **escape_utf8**: XML-encoding of unicode characters
* **xml_file_location**: Path to OCR files (when "lazy" indexing used)
* **log_dir**: Path to the log directory

#### Overview
The service works in conjunction with DSpace 7.x IIIF support. 

When indexing a new item, the service retrieves an IIIF `AnnotationList` of OCR files from the 
DSpace `Item` record. The OCR files are pre-processed based on configuration options and added to the Solr index. 
If "lazy" indexing is used, OCR files are written to disk.

Processing order is determined either by structural metadata (e.g. METS) or the order of OCR files in the DSpace bundle. 

This service can be ran on the same host as Solr to support "lazy" indexing. If you are using "full" indexing
or providing a shared file system by other means the service can run on a separate host


#### External Requirements
You must add the solr-ocrhighlighting plugin to Solr. See the instructions: https://dbmdz.github.io/solr-ocrhighlighting/installation/

You need an IIIF-enabled DSpace instance. Your DSpace `Items` must be individually enabled for IIIF and search via 
the metadata fields `dspace-iiif-enabled` and `iiif-search-enabled`. The Item's OCR files must be
in the DSpace Item's `OtherContent` Bundle. If your processing order is determined by structural metadata, be sure
to name your structural metadata file `mets.xml`. If this file does not exist or has not been correctly named, 
processing order is determined by the order of OCR files in the `OtherContent` Bundle.

See DSpace IIIF documentation: https://wiki.lyrasis.org/display/DSDOC7x/IIIF+Configuration

## Installation

#### Binary Files:

DSpace 7.x should eventually include OS-specific directories with starter configuration files and a Solr core that's pre-configured for the `solr-ocrhighlighting` plugin.

In the meantime, you can build from source.

`go build -o /output/directory main.go `

For a specific platform:

`env GOOS=<target-OS> GOARCH=<target-architecture> go build -o /output/directory main.go `

#### Docker

Pull from Docker Hub:

`docker pull mspalti/ocr_processor:latest`

Example of running the container with volumes (Linux).

`docker run -d -user <host_user_GID> --network host -v /host/path/to/configs:/processor/configs -v /host/path/to/logs:/var/log/ocr_processor -v /path/escaped/alto/files:/var/ocr_files mspalti/ocrprocessor`

Note that you don't need to create a volume for the `/var/ocr_files` mount point if you aren't using "lazy" indexing. 

If using SELinux security you may need to add `:Z` to your mount point paths, e.g. `/indexer/logs:Z`

On MacOS or Windows you can't use the `--network host` option. Instead, change DSpace and Solr URL's in 
`config.yml` to use the IP address of the host system rather than `localhost`. This appoach works only for "full"
indexing. 


## Usage

POST, DELETE, or GET requests use the identifier of a DSpace Item as follows: 

`http://<host>:3000/item/413065ef-e242-4d0e-867d-8e2f6486be56`

### DSpace command line tool (under development)

A DSpace CLI tool is currently being considered. That tool uses this service to add or delete OCR from the
Solr index. The tool allows batch updates at the Community or Collection levels, as well as individual Item 
updates. 

Usage:

**Add:**
./bin/dspace iiif-search-index --add -e user@dspace.edu -i f797f6ee-f27f-4548-8590-45d6df8a7431

**Delete:**
./bin/dspace iiif-search-index --delete -e user@dspace.edu -i f797f6ee-f27f-4548-8590-45d6df8a7431



