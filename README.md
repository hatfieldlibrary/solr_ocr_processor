# Solr Word Highlight Indexer

This prototype works with a proposed DSpace 7 IIIF implementation that is currently under review: https://github.com/DSpace/DSpace/pull/3210.  

Requires this Solr plugin: https://github.com/dbmdz/solr-ocrhighlighting.

The Go program runs on the file system used by Solr. It retrieves METS and ALTO files from DSpace using Manifests retrieved via the IIIF REST API. ALTO files are pre-processed before indexing by Solr. Solr retrieves word highlight 
coordinates from ALTO files located on the shared file system.



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
