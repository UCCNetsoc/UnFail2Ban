package main

import (
	"fmt"
	"io/ioutil"

	"github.com/BurntSushi/toml"
)

type config struct {
	Jail       string `toml:"jail"`
	Port       string `toml:"port"`
	CookieHost string `toml:"cookie_host"`
	ListenHost string `toml:"listen_host"`
	FileDir    string `toml:"file_dir"`

	LDAPKey    string `toml:"LDAP_Key"`
	LDAPHost   string `toml:"LDAP_Host"`
	LDAPUser   string `toml:"LDAP_User"`
	LDAPBaseDN string `toml:"LDAP_BaseDN"`
}

func loadConfig() error {
	confRead, err := ioutil.ReadFile("/etc/unfail2ban/settings.conf")
	if err != nil {
		return fmt.Errorf("Failed to read config file: %v", err)
	}

	if _, err = toml.Decode(string(confRead), conf); err != nil {
		return fmt.Errorf("Failed to unmarshall config: %v", err)
	}
	return nil
}
