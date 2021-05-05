package main

import (
	"fmt"
	"github.com/spf13/viper"
	"go_alto_indexer/internal"
	"os"
)

type Configuration struct {
	DSpaceHost string
	Collections []string
}

func getConfig() Configuration {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil { // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	settings := Configuration {
		DSpaceHost: viper.GetString("dspace_host"),
		Collections: viper.GetStringSlice("Collections") }
	return settings
}

func main() {
	args := os.Args[1:]
	action := ""
	item := ""
	if len(args) > 0 {
		action = args[0]
	}
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
		internal.AddToIndex(settings.DSpaceHost, item)
	}
}

