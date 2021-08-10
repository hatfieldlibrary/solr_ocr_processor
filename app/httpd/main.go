package main

import (
	"errors"
	. "github.com/mspalti/altoindexer/src/app/index"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"os"
	"strings"
)

const configFilePath = "~/configs"

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
			handleError(errors.New("missing parameter"), response, 400)
			return
		}
		itemId := pathParams[0]
		action := pathParams[1]
		// add and delete actions
		var idx Indexer
		if action == "add" {
			idx = AddItem{}
		}
		if action == "delete" {
			idx = DeleteItem{}
		}
		if idx != nil {
			err := HandleAction(idx, config, &itemId)
			if err != nil {
				handleError(err, response, 500)
				return
			}
		} else {
			handleError(errors.New("invalid action"), response, 400)
			return
		}
		response.WriteHeader(200)
		return
	}
}

func handleError(err error, response http.ResponseWriter, code int) {
		log.Println(err)
		response.WriteHeader(code)
}

func main() {
	config, err := config()
	file, err := os.OpenFile(config.LogDir + "/alto_indexer.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	log.SetOutput(file)
	if err != nil {
		log.Println(err)
		return
	}
	// TODO implement post and delete methods
	http.HandleFunc("/", configuredHandler(config))
	log.Fatal(http.ListenAndServe(":" + config.HttpPort, nil))
}
