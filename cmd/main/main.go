package main

import (
	"fmt"
	"github.com/spf13/viper"
	"go_alto_indexer/internal"
	"os"
)

func getConfig() internal.Configuration {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil { // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	settings := internal.Configuration {
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
	settings := getConfig()
	fmt.Println(settings.DSpaceHost)
	if len(settings.Collections) > 0 {
		fmt.Println(settings.Collections[0])
	} else {
		fmt.Println("No dspace collection handles provided in the configuration.")
	}
	if action == "add" && len(item) > 0 {
		internal.AddToIndex(settings, item)
	}
}

