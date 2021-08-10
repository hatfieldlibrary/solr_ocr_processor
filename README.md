# Solr Word Highlight Indexer

Note: This documentation is no longer accurate. 

This prototype works with a proposed DSpace 7 IIIF implementation that is currently under review: https://github.com/DSpace/DSpace/pull/3210.  

Requires this Solr plugin from the MDZ Digital Library team: https://github.com/dbmdz/solr-ocrhighlighting.

The Go program runs on the same file system as the Solr index. It retrieves METS and ALTO files from DSpace,
 using the IIIF manifest's seeAlso AnnotationList. ALTO files are retrieved from DSpace and 
pre-processed before indexing by Solr. 

```  
Usage:

  -action string
        the action to perform (e.g. add)
  -config string
        path to the directory that contains your config.yaml file (default "./configs")
  -item string
        the dspace item uuid
        
  Example: ./altoindexer -action=add -config=/home/user -item=f3b10302-xxxx-xxxx-xxxx-3cc3ea89b366
```
This starter implementation processes single items retrieved via their DSpace Item id.  It will need to called by 
a parent process that monitors DSpace Collections and syncs the Solr index as Items are added.
