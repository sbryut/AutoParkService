package handlers

import (
	"html/template"
	"net/http"
)

// Отображение главной рабочей страницы
func DashboardPage(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")

	userID, ok := session.Values["user_id"].(int)
	username, _ := session.Values["username"].(string)

	if !ok || userID == 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	tmpl, err := template.ParseFiles(
		"./ui/template/layout.html", "./ui/template/dashboard.html",
	)
	if err != nil {
		http.Error(w, "Error loading page", http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, struct {
		Title    string
		Username string
	}{
		Title:    "Главная",
		Username: username,
	})
}
