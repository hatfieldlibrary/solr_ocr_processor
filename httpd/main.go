package main

import (
	"errors"
	. "github.com/mspalti/alto_indexer/index"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"strings"
)

func getConfig(configFilePath string) (Configuration, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configFilePath)
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil { // Handle errors reading the config file
		return Configuration{}, errors.New("fatal error reading config file")
	}
	settings := Configuration{
		DSpaceHost: viper.GetString("dspace_host"),
		Collections: viper.GetStringSlice("Collections"),
		SolrUrl: viper.GetString("solr_url"),
		SolrCore: viper.GetString("solr_core"),
		XmlFileLocation: viper.GetString("xml_file_location"),
		LogDir: viper.GetString("log_dir"),
	}
	return settings, nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	pathParams := strings.Split(r.URL.Path, "/")[1:]
	if len(pathParams) != 2 {
		w.WriteHeader(400)
		return
	}
	configFilePath := "./configs"
	settings, err := getConfig(configFilePath)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	itemId := pathParams[0]
	action := pathParams[1]
	err = HandleAction(settings, &itemId, &action)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
	return
}

func main() {
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
