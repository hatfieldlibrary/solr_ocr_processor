# Solr Word Highlight Indexer

This prototype works with a proposed DSpace 7 IIIF implementation that is currently under review: https://github.com/DSpace/DSpace/pull/3210.  

The Go program must run on the file system used by Solr. It retrieves METS and ALTO files 
from DSpace using Manifests retrieved via the IIIF REST API. ALTO files are modifed before indexing. Word highlight coordinates are retrieved from 
ALTO files located on the shared file system.

Requires this Solr plugin: https://github.com/dbmdz/solr-ocrhighlighting.

```  
Usage:

  -action string
        the action to perform (e.g. add)
  -config string
        path to the directory that containsyour config.yaml file (default "./configs")
  -item string
        the dspace item uuid
        
  Example: ./altoindexer -action=add -config=/home/user -item=f3b10302-xxxx-xxxx-xxxx-3cc3ea89b366
```
