package controllers

import (
	"html/template"
	"net/http"

	"martins-pocos/config"
	"martins-pocos/models"
)

type AuthController struct {
	UserModel *models.UserModel
}

func NewAuthController(userModel *models.UserModel) *AuthController {
	return &AuthController{UserModel: userModel}
}

func (c *AuthController) LoginPage(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/login.html"))
	tmpl.Execute(w, nil)
}

func (c *AuthController) RegisterPage(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/register.html"))
	tmpl.Execute(w, nil)
}

func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	user, err := c.UserModel.GetByEmail(email)
	if err != nil {
		http.Error(w, "Usuário não encontrado", http.StatusUnauthorized)
		return
	}

	if !c.UserModel.ValidatePassword(password, user.Password) {
		http.Error(w, "Senha incorreta", http.StatusUnauthorized)
		return
	}

	// Create session
	session, _ := config.GetSessionStore().Get(r, "session")
	session.Values["user_id"] = user.ID
	session.Values["user_type"] = user.UserType
	session.Values["user_name"] = user.Name
	session.Save(r, w)

	// Redirect based on user type
	if user.UserType == "gestor" {
		http.Redirect(w, r, "/dashboard/admin", http.StatusFound)
	} else {
		http.Redirect(w, r, "/dashboard/cliente", http.StatusFound)
	}
}

func (c *AuthController) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := &models.User{
		Name:     r.FormValue("name"),
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
		Phone:    r.FormValue("phone"),
		Address:  r.FormValue("address"),
	}

	err := c.UserModel.Create(user)
	if err != nil {
		http.Error(w, "Erro ao criar usuário", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/login?success=1", http.StatusFound)
}

func (c *AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := config.GetSessionStore().Get(r, "session")
	session.Values["user_id"] = nil
	session.Values["user_type"] = nil
	session.Values["user_name"] = nil
	session.Options.MaxAge = -1
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)
}