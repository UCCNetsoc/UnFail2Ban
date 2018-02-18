package main

import (
	"io/ioutil"

	"github.com/BurntSushi/toml"
)

type config struct {
	Jail       string `toml:"jail"`
	Port       string `toml:"port"`
	CookieHost string `toml:"cookie_host"`
	ListenHost string `toml:"listen_host"`

	LDAPKey    string `toml:"LDAP_Key"`
	LDAPHost   string `toml:"LDAP_Host"`
	LDAPUser   string `toml:"LDAP_User"`
	LDAPBaseDN string `toml:"LDAP_BaseDN"`
}

func loadConfig() {
	confRead, err := ioutil.ReadFile("settings.conf")
	if err != nil {
		errorLog.Fatalln("Error reading config file:", err)
	}

	if _, err = toml.Decode(string(confRead), conf); err != nil {
		errorLog.Fatalln("Error unmarshalling config:", err)
	}
}
