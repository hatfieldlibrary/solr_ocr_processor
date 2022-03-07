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
// this a relative path to the local directory.
//const configFilePath = "/indexer/configs"
const configFilePath = "./configs"

func config() (*Configuration, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configFilePath)

	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		return &Configuration{}, errors.New("fatal error reading config file" + err.Error())
	}
	config := Configuration{
		DSpaceHost:      viper.GetString("dspace_host"),
		Collections:     viper.GetStringSlice("Collections"),
		SolrUrl:         viper.GetString("solr_url"),
		SolrCore:        viper.GetString("solr_core"),
		FileFormat:      viper.GetString("file_format"),
		IndexType:       viper.GetString("index_type"),
		XmlFileLocation: viper.GetString("xml_file_location"),
		HttpPort:        viper.GetString("http_port"),
		LogDir:          viper.GetString("log_dir"),
	}

	return &config, nil
}

func indexingHandler(config *Configuration) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		pathParams := strings.Split(request.URL.Path, "/")[1:]
		if !(len(pathParams) >= 3) {
			handleError(errors.New("missing parameter"), response, 400)
			return
		}
		itemId := pathParams[1]
		action := pathParams[2]

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
			handleError(errors.New("invalid or missing action"), response, 400)
			return
		}
		response.WriteHeader(200)
		return
	}
}

func handleError(err error, response http.ResponseWriter, code int) {
	log.Println(err)
	switch err.(type) {
	case UnProcessableEntity:
		response.WriteHeader(422)
	case BadRequest:
		response.WriteHeader(400)
	case MethodNotAllowed:
		response.WriteHeader(405)
	default:
		response.WriteHeader(code)
	}
}

func getLogFile(config *Configuration) (*os.File, error) {
	path := config.LogDir + "/alto_indexer.log"
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE, 0666)
	return file, err
}

func main() {

	// app configuration
	config, err := config()
	if err != nil {
		println("Server config is missing: " + err.Error())
		return
	}

	// logging
	file, err := getLogFile(config)
	if err != nil {
		return
	}
	log.SetOutput(file)


	// set up the server and handler(s)
	mux := http.NewServeMux()
	indexer := indexingHandler(config)

	// define routes
	// TODO implement post and delete
	mux.Handle("/item/", indexer)

	// listen
	serverError := http.ListenAndServe(":"+config.HttpPort, mux)
	if serverError != nil {
		log.Fatal(serverError)
	}

}
