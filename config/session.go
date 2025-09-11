package config

import (
	"github.com/gorilla/sessions"
)

var store *sessions.CookieStore

func InitSession() {
	store = sessions.NewCookieStore([]byte("martins-pocos-secret-key"))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
	}
}

func GetSessionStore() *sessions.CookieStore {
	return store
}