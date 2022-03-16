
## Processing Service for the DSpace IIIF Search API
This service pre-processes `DSpace` METS/ALTO files for indexing by the `solr-ocrhighlighting` Solr plugin. 

**DSpace**: https://wiki.lyrasis.org/display/DSDOC7x

**solr-ocrhighlighting plugin**: https://github.com/dbmdz/solr-ocrhighlighting. 

#### Supports
* GET, POST, and DELETE methods
* For POST's, MiniOcr or ALTO files are added to the index with "full" or "lazy" indexing and optional XML-encoding of Unicode characters.
* GET requests verify that OCR files have been indexed.
* DELETE requests remove OCR files from the index and the file system (if "lazy" indexing was used).

#### Configuration Options
* **http_port**: listen port of service
* **ip_whitelist**: IPs that are allowed access
* **dspace_host**: Base URL of the DSpace service
* **solr_url**: Base URL of the Solr service
* **solr_core**: Solr core ("word_highlighting")
* **file_format**: MiniOcr or ALTO
* **index_type**: Full or lazy
* **escape_utf8**: XML-encoding of unicode characters
* **xml_file_location**: Path to OCR files (when "lazy" indexing used)
* **log_dir**: Path to the log directory

#### Overview
The service works in conjunction with DSpace 7.x IIIF support. 

When indexing a new item, the service retrieves an IIIF `AnnotationList` of METS and ALTO files from the 
DSpace `Item` record. ALTO files are first pre-processed based on configuration options and then added to the Solr index. 
If "lazy" indexing is used, OCR files are written to disk.


#### External Requirements
You need an IIIF-enabled DSpace instance and DSpace `Items` that are IIIF and search-enabled via the metadata fields
`dspace-iiif-enabled` and `iiif-search-enabled`. To be available in an IIIF `AnnotationList`, METS/ALTO files must be
in the DSpace Item's `OTHER_CONTENT` Bundle.

You also need to add the solr-ocrhighlighting plugin to Solr.


## Installation

#### Binary Files:

DSpace 7.x should eventually include OS-specific directories with starter configuration files and a Solr core that's pre-configured for the `solr-ocrhighlighting` plugin.

In the meantime, you can build from source.

`go build main.go -o /output/directory`

#### Docker

Pull from Docker Hub:

`docker pull mspalti/altoindexer:latest`

Example of running the container with volumes (Linux).

`docker run -d --network host -v /host/path/to/configs:/indexer/configs -v /host/path/to/logs:/indexer/logs -v /path/escaped/alto/files:/var/ocr_files mspalti/altoindexer`

Note that you don't need to create a volume for the `/var/ocr_files` mount point if you aren't using "lazy" indexing. 

If using SELinux security you may need to add `:Z` to your mount point paths, e.g. `/indexer/logs:Z`

On MacOS or Windows you can't use the `--network host` option. Instead, change DSpace and Solr URL's in 
`config.yml` to use the IP address of the host system rather than `localhost`. This appoach works only for "full"
indexing. 


## Usage

POST, DELETE, or GET requests use the identifier of a DSpace Community, Collection or Item as follows: 

`http://<host>:3000/item/413065ef-e242-4d0e-867d-8e2f6486be56`

### DSpace command line tool (under development)

**Add:**
./bin/dspace iiif-search-index --add -e user@dspace.edu -i f797f6ee-f27f-4548-8590-45d6df8a7431

**Delete:**
./bin/dspace iiif-search-index --delete -e user@dspace.edu -i f797f6ee-f27f-4548-8590-45d6df8a7431



