package config

import (
	"encoding/json"
	"strings"

	"github.com/Strum355/log"

	"github.com/spf13/viper"
)

func Load() error {
	InitDefaults()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	return nil
}

func PrintSettings() {
	// Print settings with secrets redacted
	settings := viper.AllSettings()
	settings["windlass"].(map[string]interface{})["secret"] = "[redacted]"

	out, _ := json.MarshalIndent(settings, "", "\t")
	log.Debug("config:\n" + string(out))
}
