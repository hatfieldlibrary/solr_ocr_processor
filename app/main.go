package main

import (
	"errors"
	. "github.com/mspalti/altoindexer/index"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"os"
	"strings"
)

// This absolute path is the mount point for the
// container volume. If you are running this
// locally or not using a container, make
// this a relative path.
const configFilePath = "/app/configs"
// This absolute path is the container mount point for the log
// directory. If you change it during development be sure
// to revert to this path before pushing a container image.
const logDirectory = "/app/logs"

func config() (*Configuration, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configFilePath)

	err := viper.ReadInConfig() // Find and read the config file
	if err != nil { // Handle errors reading the config file
		return &Configuration{}, errors.New("fatal error reading config file" + err.Error())
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
		log.Println(itemId)
		log.Println(action)
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

func getLogFile() (*os.File, error) {
	path := logDirectory + "/alto_indexer.log"
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE, 0)
	return file, err
}

func main() {
	config, err := config()
	if err != nil {
		log.Println(err)
		return
	}
	file, err := getLogFile()
	if err != nil {
		log.Println(err)
		return
	}
	log.SetOutput(file)
	// TODO implement post and delete methods
	http.HandleFunc("/", configuredHandler(config))
	log.Fatal(http.ListenAndServe(":" + config.HttpPort, nil))
}
