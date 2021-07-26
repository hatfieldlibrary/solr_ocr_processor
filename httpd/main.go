package main

import (
	"errors"
	. "github.com/mspalti/alto_indexer/index"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"strings"
)

const configFilePath = "./configs"

func config() (*Configuration, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configFilePath)
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil { // Handle errors reading the config file
		return &Configuration{}, errors.New("fatal error reading config file")
	}
	config := Configuration{
		DSpaceHost: viper.GetString("dspace_host"),
		Collections: viper.GetStringSlice("Collections"),
		SolrUrl: viper.GetString("solr_url"),
		SolrCore: viper.GetString("solr_core"),
		XmlFileLocation: viper.GetString("xml_file_location"),
		HttpPort: viper.GetString("http_port"),
		LogDir: viper.GetString("log_dir"),
	}

	return &config, nil
}

func configuredHandler(config *Configuration) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		pathParams := strings.Split(request.URL.Path, "/")[1:]
		if len(pathParams) != 2 {
			response.WriteHeader(400)
			return
		}
		itemId := pathParams[0]
		action := pathParams[1]
		err := HandleAction(config, &itemId, &action)
		if err != nil {
			response.WriteHeader(500)
			return
		}
		response.WriteHeader(200)
		return
	}
}


func main() {
	config, err := config()
	if err != nil {
		return
	}
	http.HandleFunc("/", configuredHandler(config))
	log.Fatal(http.ListenAndServe(":" + config.HttpPort, nil))
}
