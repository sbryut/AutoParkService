package transport

import (
	"github.com/gorilla/mux"
	"net/http"

	"AutoParkWeb/internal/services"
	"AutoParkWeb/internal/transport/handlers"
)

func MethodOverride(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			method := r.FormValue("_method")
			if method != "" {
				r.Method = method
			}
		}
		next.ServeHTTP(w, r)
	})
}

func SetupRoutes(service *services.AutoParkService) *mux.Router {
	router := mux.NewRouter()

	router.Use(MethodOverride)

	// Создаем HTTP обработчики
	handler := handlers.NewAutoParkHandler(service)

	// Обслуживание статических файлов из ui/static
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static", http.FileServer(http.Dir("./ui/static/"))))

	// Главная страница
	router.HandleFunc("/", handlers.HomeHandler).Methods(http.MethodGet)

	// Маршрут для страницы логина
	router.HandleFunc("/login", handlers.LoginPage(service)).Methods(http.MethodGet, http.MethodPost)
	// Маршрут для регистрации
	router.HandleFunc("/register", handlers.RegisterPage(service)).Methods(http.MethodGet, http.MethodPost)

	// Маршрут для рабочей страницы
	router.HandleFunc("/dashboard", handlers.DashboardPage).Methods(http.MethodGet)

	// Маршруты для работы с водителями
	router.HandleFunc("/drivers", handler.GetDrivers).Methods(http.MethodGet)
	router.HandleFunc("/drivers/new", handler.AddDriverPage).Methods(http.MethodGet)
	router.HandleFunc("/drivers", handler.AddDriver).Methods(http.MethodPost)
	router.HandleFunc("/drivers/{id}/edit", handler.EditDriverPage).Methods(http.MethodGet)
	router.HandleFunc("/drivers/{id}", handler.UpdateDriver).Methods(http.MethodPost)
	router.HandleFunc("/drivers/{id}/delete", handler.DeleteDriver).Methods(http.MethodPost)

	// Маршруты для работы с автомобилями
	router.HandleFunc("/autos", handler.GetCars).Methods(http.MethodGet)
	router.HandleFunc("/autos/new", handler.AddCarPage).Methods(http.MethodGet)
	router.HandleFunc("/autos", handler.AddCar).Methods(http.MethodPost)
	router.HandleFunc("/autos/{id}/edit", handler.EditCarPage).Methods(http.MethodGet)
	router.HandleFunc("/autos/{id}", handler.UpdateCar).Methods(http.MethodPost)
	router.HandleFunc("/autos/{id}/delete", handler.DeleteCar).Methods(http.MethodPost)

	// Маршруты для работы с маршрутами
	router.HandleFunc("/routes", handler.GetRoutes).Methods(http.MethodGet)
	router.HandleFunc("/routes/new", handler.AddRoutePage).Methods(http.MethodGet)
	router.HandleFunc("/routes", handler.AddRoute).Methods(http.MethodPost)
	router.HandleFunc("/routes/{id}/edit", handler.EditRoutePage).Methods(http.MethodGet)
	router.HandleFunc("/routes/{id}", handler.UpdateRoute).Methods(http.MethodPost)
	router.HandleFunc("/routes/{id}/delete", handler.DeleteRoute).Methods(http.MethodPost)

	// Маршруты для работы с журналом
	router.HandleFunc("/download", handler.DownloadJournal).Methods(http.MethodGet)
	router.HandleFunc("/journal", handler.GetAllJournalEntries).Methods(http.MethodGet)
	router.HandleFunc("/journal/new", handler.AddJournalEntryPage).Methods(http.MethodGet)
	router.HandleFunc("/journal", handler.AddJournalEntry).Methods(http.MethodPost)
	router.HandleFunc("/journal/{id}/edit", handler.EditJournalEntryPage).Methods(http.MethodGet)
	router.HandleFunc("/journal/{id}/edit", handler.EditJournalEntryPage).Methods(http.MethodGet)
	router.HandleFunc("/journal/{id}/complete", handler.CompleteJournalEntry).Methods(http.MethodPost)
	router.HandleFunc("/journal/{id}/delete", handler.DeleteJournalEntry).Methods(http.MethodPost)
	router.HandleFunc("/journal/{id}/update", handler.UpdateJournalEntry).Methods(http.MethodPost)

	// Процедуры для аналитики
	router.HandleFunc("/statistics", handler.StatisticsPage).Methods(http.MethodGet)

	return router
}
