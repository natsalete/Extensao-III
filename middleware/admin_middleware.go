package middleware

import (
	"net/http"

	"martins-pocos/config"
)

// AdminMiddleware verifica se o usuário é administrador/gestor
func AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := config.GetSessionStore().Get(r, "session")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		// Verificar se está autenticado
		authenticated, ok := session.Values["authenticated"].(bool)
		if !ok || !authenticated {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		// Verificar se é admin/gestor (user_type_id = 2)
		userTypeID, ok := session.Values["user_type_id"].(int)
		if !ok || userTypeID != 2 {
			http.Error(w, "Acesso negado. Área restrita para administradores.", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// AuthMiddleware verifica se o usuário está autenticado
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := config.GetSessionStore().Get(r, "session")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		authenticated, ok := session.Values["authenticated"].(bool)
		if !ok || !authenticated {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}