package main

import (
	"io/ioutil"

	"github.com/BurntSushi/toml"
)

func loadConfig() {
	confRead, err := ioutil.ReadFile("settings.conf")
	if err != nil {
		errorLog.Fatalln("Error reading config file:", err)
	}

	if _, err = toml.Decode(string(confRead), conf); err != nil {
		errorLog.Fatalln("Error unmarshalling config:", err)
	}
}
