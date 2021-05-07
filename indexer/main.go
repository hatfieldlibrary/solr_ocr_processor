package main

import (
	"flag"
	"fmt"
	"github.com/spf13/viper"
	"os"
)

func getConfig(configFilePath *string) Configuration {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(*configFilePath)
	print(*configFilePath)
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil { // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	settings := Configuration{
		DSpaceHost: viper.GetString("dspace_host"),
		Collections: viper.GetStringSlice("Collections"),
		SolrUrl: viper.GetString("solr_url"),
		SolrCore: viper.GetString("solr_core"),
		XmlFileLocation: viper.GetString("xml_file_location"),
		LogDir: viper.GetString("log_dir"),
	}
	return settings
}

func main() {

	configFilePath := flag.String("config", "./configs", "path to the directory that contains" +
		"your config.yaml file")

	args := os.Args[1:]
	action := ""
	item := ""
	// action
	if len(args) > 0 {
		action = args[0]
	}
	// item uuid
	if len(args) > 1 {
		item = args[1]
	}
	settings := getConfig(configFilePath)
	fmt.Println(settings.DSpaceHost)
	if len(settings.Collections) > 0 {
		fmt.Println(settings.Collections[0])
	} else {
		fmt.Println("No dspace collection handles provided in the configuration.")
	}
	if action == "add" && len(item) > 0 {
		AddToIndex(settings, item)
	}
}

