package middleware

import (
	"net/http"

	"martins-pocos/config"
)

func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := config.GetSessionStore().Get(r, "session")
		
		userID, ok := session.Values["user_id"]
		if !ok || userID == nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		
		next(w, r)
	}
}

func RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := config.GetSessionStore().Get(r, "session")
		
		userType, ok := session.Values["user_type"].(string)
		if !ok || userType != "gestor" {
			http.Error(w, "Acesso negado", http.StatusForbidden)
			return
		}
		
		next(w, r)
	}
}

func RequireClient(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := config.GetSessionStore().Get(r, "session")
		
		userType, ok := session.Values["user_type"].(string)
		if !ok || userType != "cliente" {
			http.Error(w, "Acesso negado", http.StatusForbidden)
			return
		}
		
		next(w, r)
	}
}