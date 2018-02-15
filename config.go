package main

import (
	"github.com/BurntSushi/toml"
	"io/ioutil"
)

func loadConfig() {
	confRead, err := ioutil.ReadFile("settings.conf")
	if err != nil {
		errorLog.Fatalln("Error reading config file:", err)
	}

	_, err = toml.Decode(string(confRead), conf)
	if err != nil {
		errorLog.Fatalln("Error unmarshalling config:", err)
	}
}
