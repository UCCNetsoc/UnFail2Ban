package main

import (
	"fmt"
	"net/http"
)

func checkCookie(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := store.Get(r, "id")
		if err != nil {
			http.Redirect(w, r, "/?auth=err", http.StatusTemporaryRedirect)
			errorLog.Printf("Failed to get session from store: %v", err)
			return
		}

		if session.IsNew {
			http.Redirect(w, r, "/?auth=nologin", http.StatusTemporaryRedirect)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getUserFromSession(r *http.Request) (user, error) {
	session, err := store.Get(r, "id")
	if err != nil {
		errorLog.Printf("Failed to get session from store: %v", err)
		return user{}, err
	}

	if val, ok := session.Values["user"]; ok {
		if u, ok := val.(user); ok {
			return u, nil
		}
		return user{}, fmt.Errorf("value not a user %v", val)
	}

	return user{}, fmt.Errorf("no value found")
}
