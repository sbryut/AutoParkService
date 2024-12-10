package handlers

import (
	"github.com/gorilla/sessions"
	"html/template"
	"log"
	"net/http"

	"AutoParkWeb/internal/services"
)

var store = sessions.NewCookieStore([]byte("your-secret-key"))

func init() {
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   3600, // Время жизни сессии = 1 час
		HttpOnly: true,
		Secure:   true,
	}
}

func LoginPage(service *services.AutoParkService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			username := r.FormValue("username")
			password := r.FormValue("password")

			user, err := service.AuthenticateUser(username, password)
			if err != nil {
				http.Error(w, "Invalid credentials", http.StatusUnauthorized)
				return
			}

			session, _ := store.Get(r, "session-name")

			session.Values["user_id"] = user.ID
			session.Values["username"] = user.Username
			session.Values["user_role"] = user.Role
			session.Save(r, w)

			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}

		http.ServeFile(w, r, "ui/template/login.html")
	}
}
func RegisterPage(service *services.AutoParkService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			tmpl, err := template.ParseFiles("ui/template/register.html")
			if err != nil {
				log.Printf("Ошибка парсинга шаблона: %v", err)
				http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
				return
			}
			tmpl.Execute(w, nil)
			return
		}

		if r.Method == http.MethodPost {
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Ошибка обработки формы", http.StatusBadRequest)
				return
			}

			username := r.Form.Get("username")
			password := r.Form.Get("password")
			confirmPassword := r.Form.Get("confirm_password")

			if password != confirmPassword {
				http.Error(w, "Пароли не совпадают", http.StatusBadRequest)
				return
			}

			err := service.RegisterUser(username, password)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			user, err := service.AuthenticateUser(username, password)
			if err != nil {
				http.Error(w, "Ошибка аутентификации", http.StatusInternalServerError)
				return
			}

			session, _ := store.Get(r, "session-name")
			session.Values["user_id"] = user.ID
			session.Values["username"] = user.Username
			session.Values["user_role"] = user.Role
			session.Save(r, w)

			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		}
	}
}
