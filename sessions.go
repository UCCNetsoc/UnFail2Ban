package main

import (
	"github.com/gorilla/sessions"
	"net/http"
)

var store = sessions.NewCookieStore([]byte("1234567890123456"))

func init() {
	store.Options = &sessions.Options{
		Domain:   "127.0.0.1",
		MaxAge:   60 * 10,
		HttpOnly: true,
	}
}

func checkCookie(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("id")
		if err == http.ErrNoCookie {
			w.Header().Set("auth", "nologin")
			http.Redirect(w, r, "/", http.StatusUnauthorized)
			return
		}

		session, err := store.Get(r, cookie.Value)
		if err != nil {
			w.Header().Set("auth", "err")
			http.Redirect(w, r, "/", http.StatusInternalServerError)
			return
		}

		if session.IsNew {
			w.Header().Set("auth", "nologin")
			http.Redirect(w, r, "/", http.StatusUnauthorized)
			return
		}

		if val, ok := session.Values["admin"]; ok {
			if val.(bool) {
				http.Redirect(w, r, "/list", http.StatusOK)
				return
			}
			http.Redirect(w, r, "/noauth", http.StatusUnauthorized)
			return
		}
	})
}
