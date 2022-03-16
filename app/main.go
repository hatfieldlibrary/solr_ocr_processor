package main

import (
	"errors"
	. "github.com/mspalti/altoindexer/index"
	"github.com/spf13/viper"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
)

// This absolute path is the mount point for the
// container volume. If you are not running inside a
// container use a relative path to the local configuration
// directory.
// Container mount point:
// const configFilePath = "/indexer/configs"
// Local configuration directory
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
		EscapeUtf8:      viper.GetBool("escape_utf8"),
		XmlFileLocation: viper.GetString("xml_file_location"),
		HttpPort:        viper.GetString("http_port"),
		IpWhitelist:     viper.GetStringSlice("ip_whitelist"),
		LogDir:          viper.GetString("log_dir"),
	}

	return &config, nil
}

// checkWhitelist verify that the host is in the whitelist from configuration
func checkWhitelist(request *http.Request, whitelist []string) bool {
	ip, _, _ := net.SplitHostPort(request.RemoteAddr)
	ipString := net.ParseIP(ip).String()
	var inWhitelist = false
	println(len(whitelist))
	if len(whitelist) == 0 {
		inWhitelist = true
	}
	for _, x := range whitelist {
		if x == ipString {
			inWhitelist = true
			break
		}
	}
	return inWhitelist
}

func indexingHandler(config *Configuration, logger *log.Logger) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		// verify that the remote host is in whitelist
		inWhitelist := checkWhitelist(request, config.IpWhitelist)
		if !inWhitelist {
			handleError(errors.New("request refused because remote address is not in whitelist"),
				response, 403)
			return
		}

		// get the item identifier from the request
		pathParams := strings.Split(request.URL.Path, "/")[1:]
		if !(len(pathParams) >= 2) {
			handleError(errors.New("missing parameter"), response, 400)
			return
		}
		itemId := pathParams[1]

		// set action
		var idx Indexer
		if request.Method == "GET" {
			idx = GetItem{}
		}
		if request.Method == "POST" {
			idx = AddItem{}
		}
		if request.Method == "DELETE" {
			idx = DeleteItem{}
		}

		if idx != nil {
			err := HandleAction(idx, config, &itemId, logger)
			if err != nil {
				handleError(err, response, 500)
				return
			}
		} else {
			logger.Println("Missing or invalid processing action.")
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
	case NotFound:
		response.WriteHeader(404)
	default:
		response.WriteHeader(code)
	}
}

func getLogFile(config *Configuration) (*os.File, error) {
	path := config.LogDir + "/alto_indexer.log"
	file, err := os.OpenFile(path, os.O_RDWR | os.O_APPEND|os.O_CREATE, 0666)
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
		println("Log file directory not found: " + err.Error())
		return
	}
	defer file.Close()
	logger := log.New(file, "indexer: ", log.LstdFlags)


	// set up the server and handler(s)
	mux := http.NewServeMux()
	indexer := indexingHandler(config, logger)

	// define routes
	mux.Handle("/item/", indexer)

	// listen
	serverError := http.ListenAndServe(":"+config.HttpPort, mux)
	if serverError != nil {
		logger.Fatal(serverError)
	}

}
