package main

import (
	"errors"
	"fmt"
	"strings"

	"gopkg.in/ldap.v2"
)

var (
	errWrongPass = errors.New("wrong password")
	errNoUser    = errors.New("user not found")
)

type user struct {
	dn       string
	Username string
	Group    string
	isadmin  bool
}

func getUserFromLDAP(username, password string) (user, error) {
	u := user{Username: username}

	l, err := ldap.Dial("tcp", conf.LDAPHost)
	if err != nil {
		errorLog.Println(err)
		return u, err
	}
	defer l.Close()

	if err = l.Bind(conf.LDAPUser, conf.LDAPKey); err != nil {
		errorLog.Println(err)
		return u, err
	}

	searchRequest := ldap.NewSearchRequest(
		conf.LDAPBaseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		1, 0, false,
		fmt.Sprintf("(&(objectClass=account)(uid=%s))",
			ldap.EscapeFilter(username)),
		[]string{},
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		errorLog.Println(err)
		return u, err
	}

	if len(sr.Entries) != 1 {
		return u, errNoUser
	}

	unmarshallDNS(&u, sr.Entries[0].DN)

	if !verifyPassword(password, u, l) {
		return u, errWrongPass
	}

	return u, nil
}

func verifyPassword(pass string, u user, l *ldap.Conn) bool {
	if err := l.Bind(u.dn, pass); err != nil {
		return false
	}
	return true
}

func unmarshallDNS(u *user, dn string) {
	grouped := strings.Split(dn, ",")[:3]
	mapf(grouped, func(s string) string {
		return strings.Split(s, "=")[1]
	})

	u.dn = dn
	u.Group = grouped[1]
	u.isadmin = grouped[1] == "admins"
}
