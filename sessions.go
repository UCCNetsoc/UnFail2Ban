package main

import (
	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte("1234567890123456"))

func init() {
	store.Options = &sessions.Options{
		Domain: "127.0.0.1",
		MaxAge: 60 * 10,
	}
}
