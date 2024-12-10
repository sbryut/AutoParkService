package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/xuri/excelize/v2"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	"AutoParkWeb/internal/models"
	"AutoParkWeb/internal/services"
	"github.com/gorilla/mux"
)

type AutoParkHandler struct {
	service *services.AutoParkService
}

func NewAutoParkHandler(service *services.AutoParkService) *AutoParkHandler {
	return &AutoParkHandler{service: service}
}

// Получение роли пользователя
func (h *AutoParkHandler) getUserRole(r *http.Request) (string, error) {
	session, err := store.Get(r, "session-name")
	if err != nil {
		return "", fmt.Errorf("ошибка при получении сессии: %w", err)
	}

	userRole, ok := session.Values["user_role"].(string)
	if !ok {
		return "", fmt.Errorf("не удалось получить роль пользователя")
	}

	return userRole, nil
}

func (h *AutoParkHandler) getUserName(r *http.Request) (string, error) {
	session, err := store.Get(r, "session-name")
	if err != nil {
		return "", fmt.Errorf("ошибка при получении сессии: %w", err)
	}

	username, ok := session.Values["username"].(string)
	if !ok {
		return "", fmt.Errorf("не удалось получить имя пользователя")
	}

	return username, nil
}

// Метод для получения списка водителей
func (h *AutoParkHandler) GetDrivers(w http.ResponseWriter, r *http.Request) {
	drivers, err := h.service.GetDrivers(r.Context())
	if err != nil {
		http.Error(w, "Не удалось загрузить список водителей", http.StatusInternalServerError)
		return
	}

	userRole, err := h.getUserRole(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	userName, err := h.getUserName(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	tmpl, err := template.ParseFiles(
		"./ui/template/layout.html", "./ui/template/drivers_table/drivers.html",
	)
	if err != nil {
		http.Error(w, "Ошибка загрузки шаблона", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, struct {
		Title    string
		Drivers  []models.AutoPersonal
		UserRole string
		Username string
	}{
		Title:    "Водители",
		Drivers:  drivers,
		UserRole: userRole,
		Username: userName,
	})

	if err != nil {
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
		return
	}
}

// Форма для добавления нового водителя
func (h *AutoParkHandler) AddDriverPage(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.ParseFiles(
		"./ui/template/layout.html", "./ui/template/drivers_table/add_driver.html",
	)
	if err != nil {
		http.Error(w, "Ошибка загрузки шаблона", http.StatusInternalServerError)
		return
	}

	userName, err := h.getUserName(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	tmpl.Execute(w, struct {
		Title    string
		Username string
	}{
		Title:    "Добавление водителя",
		Username: userName,
	})
}

// Добавление водителя
func (h *AutoParkHandler) AddDriver(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		firstName := r.FormValue("first_name")
		lastName := r.FormValue("last_name")
		fatherName := r.FormValue("father_name")

		err := h.service.AddDriver(r.Context(), firstName, lastName, fatherName)
		if err != nil {
			http.Error(w, "Не удалось добавить водителя", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/drivers", http.StatusSeeOther)
	} else {
		tmpl, err := template.ParseFiles(
			"./ui/template/layout.html", "./ui/template/drivers_table/add_driver.html",
		)
		if err != nil {
			http.Error(w, "Ошибка загрузки шаблона", http.StatusInternalServerError)
			return
		}
		userName, err := h.getUserName(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		tmpl.Execute(w, struct {
			Title    string
			Username string
		}{
			Title:    "Добавить водителя",
			Username: userName,
		})
	}
}

// Страница редактирования водителя
func (h *AutoParkHandler) EditDriverPage(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}

	driver, err := h.service.GetDriverByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Не удалось загрузить данные водителя", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles(
		"./ui/template/layout.html", "./ui/template/drivers_table/edit_driver.html",
	)
	if err != nil {
		http.Error(w, "Ошибка загрузки шаблона", http.StatusInternalServerError)
		return
	}
	userName, err := h.getUserName(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	tmpl.Execute(w, struct {
		Title    string
		Driver   models.AutoPersonal
		Username string
	}{
		Title:    "Редактирование водителя",
		Driver:   *driver,
		Username: userName,
	})
}

// Обновление водителя
func (h *AutoParkHandler) UpdateDriver(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost && r.FormValue("_method") == "PUT" {
		r.Method = http.MethodPut
	}

	if r.Method == http.MethodPut {
		driverIDStr := mux.Vars(r)["id"]
		driverID, err := strconv.Atoi(driverIDStr)
		if err != nil {
			http.Error(w, "Некорректный ID водителя", http.StatusBadRequest)
			return
		}

		firstName := r.FormValue("first_name")
		lastName := r.FormValue("last_name")
		fatherName := r.FormValue("father_name")

		err = h.service.UpdateDriver(r.Context(), driverID, firstName, lastName, fatherName)
		if err != nil {
			http.Error(w, "Не удалось обновить данные водителя", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/drivers", http.StatusSeeOther)
	}
}

// Удаление водителя
func (h *AutoParkHandler) DeleteDriver(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	driverID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Invalid driver ID", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteDriver(r.Context(), driverID)
	if err != nil {
		http.Error(w, "Failed to delete driver: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Метод для получения списка автомобилей
func (h *AutoParkHandler) GetCars(w http.ResponseWriter, r *http.Request) {
	cars, err := h.service.GetCars(r.Context())
	if err != nil {
		http.Error(w, "Не удалось загрузить список автомобилей", http.StatusInternalServerError)
		return
	}

	userRole, err := h.getUserRole(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	tmpl, err := template.ParseFiles(
		"./ui/template/layout.html", "./ui/template/autos_table/autos.html",
	)
	if err != nil {
		http.Error(w, "Ошибка загрузки шаблона", http.StatusInternalServerError)
		return
	}
	userName, err := h.getUserName(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	err = tmpl.Execute(w, struct {
		Title    string
		Autos    []models.Auto
		UserRole string
		Username string
	}{
		Title:    "Автомобили",
		Autos:    cars,
		UserRole: userRole,
		Username: userName,
	})
	if err != nil {
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
		return
	}
}

// Страница для добавления нового автомобиля
func (h *AutoParkHandler) AddCarPage(w http.ResponseWriter, r *http.Request) {
	drivers, err := h.service.GetDrivers(r.Context())
	if err != nil {
		http.Error(w, "Не удалось загрузить список водителей", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles(
		"./ui/template/layout.html", "./ui/template/autos_table/add_auto.html",
	)
	if err != nil {
		http.Error(w, "Ошибка загрузки шаблона", http.StatusInternalServerError)
		return
	}
	userName, err := h.getUserName(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	err = tmpl.Execute(w, struct {
		Title    string
		Drivers  []models.AutoPersonal
		Username string
	}{
		Title:    "Добавить автомобиль",
		Drivers:  drivers,
		Username: userName,
	})
	if err != nil {
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
		return
	}
}

// Добавление нового автомобиля
func (h *AutoParkHandler) AddCar(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Получаем данные из формы
		num := r.FormValue("num")
		color := r.FormValue("color")
		mark := r.FormValue("mark")
		driverID, err := strconv.Atoi(r.FormValue("driver_id"))
		if err != nil {
			http.Error(w, "Некорректный driver_id", http.StatusBadRequest)
			return
		}

		err = h.service.AddCar(r.Context(), num, color, mark, driverID)
		if err != nil {
			http.Error(w, "Не удалось добавить автомобиль", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/autos", http.StatusSeeOther)
	} else {
		h.AddCarPage(w, r)
	}
}

// Страница редактирования автомобиля
func (h *AutoParkHandler) EditCarPage(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Некорректный ID автомобиля", http.StatusBadRequest)
		return
	}

	userRole, err := h.getUserRole(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	car, driverName, err := h.service.GetCarByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Не удалось загрузить данные автомобиля", http.StatusInternalServerError)
		return
	}

	drivers, err := h.service.GetDrivers(r.Context())
	if err != nil {
		http.Error(w, "Не удалось загрузить список водителей", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles(
		"./ui/template/layout.html", "./ui/template/autos_table/edit_auto.html",
	)
	if err != nil {
		http.Error(w, "Ошибка загрузки шаблона", http.StatusInternalServerError)
		return
	}
	userName, err := h.getUserName(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	err = tmpl.Execute(w, struct {
		Title      string
		Car        *models.Auto
		DriverName string
		Drivers    []models.AutoPersonal
		UserRole   string
		Username   string
	}{
		Title:      "Редактирование автомобиля",
		Car:        car,
		DriverName: driverName,
		Drivers:    drivers,
		UserRole:   userRole,
		Username:   userName,
	})
	if err != nil {
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
		return
	}
}

// Обновление данных автомобиля
func (h *AutoParkHandler) UpdateCar(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost && r.FormValue("_method") == "PUT" {
		r.Method = http.MethodPut
	}

	if r.Method == http.MethodPut {
		carIDStr := mux.Vars(r)["id"]
		carID, err := strconv.Atoi(carIDStr)
		if err != nil {
			http.Error(w, "Некорректный ID автомобиля", http.StatusBadRequest)
			return
		}

		num := r.FormValue("num")
		color := r.FormValue("color")
		mark := r.FormValue("mark")
		personalID, err := strconv.Atoi(r.FormValue("personal_id"))
		if err != nil {
			http.Error(w, "Некорректный personal_id", http.StatusBadRequest)
			return
		}

		if num == "" || color == "" || mark == "" {
			http.Error(w, "Все поля (num, color, mark) обязательны", http.StatusBadRequest)
			return
		}

		err = h.service.UpdateCar(r.Context(), carID, num, color, mark, personalID)
		if err != nil {
			http.Error(w, "Не удалось обновить данные автомобиля", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/autos", http.StatusSeeOther)
	}
}

// Удаление автомобиля
func (h *AutoParkHandler) DeleteCar(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	carID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Некорректный ID автомобиля", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteCar(r.Context(), carID)
	if err != nil {
		http.Error(w, "Не удалось удалить автомобиль: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Метод для получения списка маршрутов
func (h *AutoParkHandler) GetRoutes(w http.ResponseWriter, r *http.Request) {
	routes, err := h.service.GetRoutes(r.Context())
	if err != nil {
		http.Error(w, "Не удалось загрузить список маршрутов", http.StatusInternalServerError)
		return
	}

	userRole, err := h.getUserRole(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	tmpl, err := template.ParseFiles(
		"./ui/template/layout.html", "./ui/template/routes_table/routes.html",
	)
	if err != nil {
		http.Error(w, "Ошибка загрузки шаблона", http.StatusInternalServerError)
		return
	}
	userName, err := h.getUserName(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	err = tmpl.Execute(w, struct {
		Title    string
		Routes   []models.Route
		UserRole string
		Username string
	}{
		Title:    "Маршруты",
		Routes:   routes,
		UserRole: userRole,
		Username: userName,
	})

	if err != nil {
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
		return
	}
}

// Форма для добавления нового маршрута
func (h *AutoParkHandler) AddRoutePage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles(
		"./ui/template/layout.html", "./ui/template/routes_table/add_route.html",
	)
	if err != nil {
		http.Error(w, "Ошибка загрузки шаблона", http.StatusInternalServerError)
		return
	}
	userName, err := h.getUserName(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	tmpl.Execute(w, struct {
		Title    string
		Username string
	}{
		Title:    "Добавление маршрута",
		Username: userName,
	})
}

// Добавление маршрута
func (h *AutoParkHandler) AddRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		startPoint := r.FormValue("start_point")
		endPoint := r.FormValue("end_point")

		err := h.service.AddRoute(r.Context(), startPoint, endPoint)
		if err != nil {
			http.Error(w, "Не удалось добавить маршрут", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/routes", http.StatusSeeOther)
	}
}

// Форма для изменения маршрута
func (h *AutoParkHandler) EditRoutePage(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}

	route, err := h.service.GetRouteByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Не удалось загрузить данные маршрута", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles(
		"./ui/template/layout.html", "./ui/template/routes_table/edit_route.html",
	)
	if err != nil {
		http.Error(w, "Ошибка загрузки шаблона", http.StatusInternalServerError)
		return
	}
	userName, err := h.getUserName(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	tmpl.Execute(w, struct {
		Title    string
		Route    models.Route
		Username string
	}{
		Title:    "Редактирование маршрута",
		Route:    *route,
		Username: userName,
	})
}

// Обновление маршрута
func (h *AutoParkHandler) UpdateRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost && r.FormValue("_method") == "PUT" {
		r.Method = http.MethodPut
	}

	if r.Method == http.MethodPut {
		routeIDStr := mux.Vars(r)["id"]
		routeID, err := strconv.Atoi(routeIDStr)
		if err != nil {
			http.Error(w, "Некорректный ID маршрута", http.StatusBadRequest)
			return
		}

		startPoint := r.FormValue("start_point")
		endPoint := r.FormValue("end_point")

		route := &models.Route{
			ID:         routeID,
			StartPoint: startPoint,
			EndPoint:   endPoint,
		}

		err = h.service.UpdateRoute(r.Context(), route)
		if err != nil {
			http.Error(w, "Не удалось обновить данные маршрута", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/routes", http.StatusSeeOther)
	}
}

// Удаление маршрута
func (h *AutoParkHandler) DeleteRoute(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	routeID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Invalid route ID", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteRoute(r.Context(), routeID)
	if err != nil {
		http.Error(w, "Failed to delete route: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Обработчик для скачивания Excel-файла журнала
func (h *AutoParkHandler) DownloadJournal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	entries, err := h.service.GetAllJournalEntries(ctx)
	if err != nil {
		http.Error(w, "Ошибка при получении записей журнала", http.StatusInternalServerError)
		return
	}

	f := excelize.NewFile()
	sheetName := "Журнал"
	f.NewSheet(sheetName)

	f.SetCellValue(sheetName, "A1", "Маршрут")
	f.SetCellValue(sheetName, "B1", "Автомобиль")
	f.SetCellValue(sheetName, "C1", "Водитель")
	f.SetCellValue(sheetName, "D1", "Время отправления")
	f.SetCellValue(sheetName, "E1", "Время прибытия")

	for i, entry := range entries {
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", i+2), fmt.Sprintf("%s - %s", entry.StartPoint, entry.EndPoint))
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", i+2), fmt.Sprintf("%s (%s)", entry.AutoNumber, entry.AutoMark))
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", i+2), entry.DriverName)

		timeOutStr := entry.TimeOut.Format("02.01.2006 15:04")
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", i+2), timeOutStr)

		if entry.TimeIn != nil {
			timeInStr := entry.TimeIn.Format("02.01.2006 15:04")
			f.SetCellValue(sheetName, fmt.Sprintf("E%d", i+2), timeInStr)
		} else {
			f.SetCellValue(sheetName, fmt.Sprintf("E%d", i+2), "В пути")
		}
	}

	w.Header().Set("Content-Disposition", "attachment; filename=journal.xlsx")
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")

	if err := f.Write(w); err != nil {
		http.Error(w, "Ошибка при создании файла Excel", http.StatusInternalServerError)
		return
	}
}

// Получение всех записей журнала
func (h *AutoParkHandler) GetAllJournalEntries(w http.ResponseWriter, r *http.Request) {
	entries, err := h.service.GetAllJournalEntries(r.Context())
	if err != nil {
		http.Error(w, "Не удалось загрузить записи журнала", http.StatusInternalServerError)
		return
	}

	userRole, err := h.getUserRole(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	tmpl, err := template.ParseFiles(
		"./ui/template/layout.html", "./ui/template/journal_table/journal.html",
	)
	if err != nil {
		http.Error(w, "Ошибка загрузки шаблона", http.StatusInternalServerError)
		return
	}
	userName, err := h.getUserName(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	err = tmpl.Execute(w, struct {
		Title    string
		Entries  []models.JournalView
		UserRole string
		Username string
	}{
		Title:    "Записи журнала",
		Entries:  entries,
		UserRole: userRole,
		Username: userName,
	})

	if err != nil {
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
		return
	}
}

// Обработчик для получения автомобилей по водителю
func (h *AutoParkHandler) GetAutosByDriver(w http.ResponseWriter, r *http.Request) {
	driverID := r.URL.Query().Get("driver_id")
	if driverID == "" {
		http.Error(w, "Не указан ID водителя", http.StatusBadRequest)
		return
	}

	driverIDInt, err := strconv.Atoi(driverID)
	if err != nil {
		http.Error(w, "Неверный формат ID водителя", http.StatusBadRequest)
		return
	}

	autos, err := h.service.GetAutosByDriverID(r.Context(), driverIDInt)
	if err != nil {
		http.Error(w, "Не удалось загрузить автомобили", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(autos)
}

func (h *AutoParkHandler) AddJournalEntryPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	drivers, err := h.service.GetDrivers(ctx)
	if err != nil {
		log.Printf("Ошибка получения водителей: %v", err)
		http.Error(w, "Не удалось загрузить список водителей", http.StatusInternalServerError)
		return
	}

	routes, err := h.service.GetRoutes(ctx)
	if err != nil {
		log.Printf("Ошибка получения маршрутов: %v", err)
		http.Error(w, "Не удалось загрузить список маршрутов", http.StatusInternalServerError)
		return
	}

	driversAutos := make(map[int][]models.Auto)
	for _, driver := range drivers {
		driverAutos, err := h.service.GetAutosByDriverID(ctx, driver.ID)
		if err != nil {
			log.Printf("Ошибка получения авто для водителя %d: %v", driver.ID, err)
			continue
		}
		driversAutos[driver.ID] = driverAutos
	}

	layoutBytes, err := os.ReadFile("./ui/template/layout.html")
	if err != nil {
		log.Printf("Ошибка чтения layout: %v", err)
		http.Error(w, "Ошибка чтения шаблона", http.StatusInternalServerError)
		return
	}

	addJournalBytes, err := os.ReadFile("./ui/template/journal_table/add_journal.html")
	if err != nil {
		log.Printf("Ошибка чтения add_journal: %v", err)
		http.Error(w, "Ошибка чтения шаблона", http.StatusInternalServerError)
		return
	}

	funcMap := template.FuncMap{
		"json": func(v interface{}) template.JS {
			a, _ := json.Marshal(v)
			return template.JS(a)
		},
	}

	tmpl := template.New("layout").Funcs(funcMap)

	tmpl, err = tmpl.Parse(string(layoutBytes))
	if err != nil {
		log.Printf("Ошибка парсинга layout: %v", err)
		http.Error(w, "Ошибка парсинга шаблона layout", http.StatusInternalServerError)
		return
	}

	tmpl, err = tmpl.Parse(string(addJournalBytes))
	if err != nil {
		log.Printf("Ошибка парсинга add_journal: %v", err)
		http.Error(w, "Ошибка парсинга шаблона add_journal", http.StatusInternalServerError)
		return
	}
	userName, err := h.getUserName(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	data := struct {
		Title        string
		Drivers      []models.AutoPersonal
		DriversAutos map[int][]models.Auto
		Routes       []models.Route
		Username     string
	}{
		Title:        "Добавление записи в журнал",
		Drivers:      drivers,
		DriversAutos: driversAutos,
		Routes:       routes,
		Username:     userName,
	}

	err = tmpl.ExecuteTemplate(w, "layout", data)
	if err != nil {
		log.Printf("Ошибка выполнения шаблона: %v", err)
		http.Error(w, "Ошибка отображения страницы: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *AutoParkHandler) AddJournalEntry(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Ошибка парсинга формы", http.StatusBadRequest)
		return
	}

	driverID, _ := strconv.Atoi(r.Form.Get("driver_id"))
	autoID, _ := strconv.Atoi(r.Form.Get("auto_id"))
	routeID, _ := strconv.Atoi(r.Form.Get("route_id"))
	timeOut := r.Form.Get("time_out")

	if driverID == 0 || autoID == 0 || routeID == 0 || timeOut == "" {
		http.Error(w, "Не все поля заполнены", http.StatusBadRequest)
		return
	}

	err := h.service.AddJournalEntry(r.Context(), autoID, routeID, timeOut)
	if err != nil {
		log.Printf("Ошибка добавления записи в журнал: %v", err)
		http.Error(w, "Не удалось добавить запись", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/journal", http.StatusSeeOther)
}

// Страница для редактирования записи журнала
func (h *AutoParkHandler) EditJournalEntryPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Некорректный ID записи журнала", http.StatusBadRequest)
		return
	}

	entry, err := h.service.GetJournalEntryByID(ctx, id)
	if err != nil {
		log.Printf("Ошибка получения записи журнала: %v", err)
		http.Error(w, "Не удалось загрузить данные записи журнала", http.StatusInternalServerError)
		return
	}

	drivers, err := h.service.GetDrivers(ctx)
	if err != nil {
		log.Printf("Ошибка получения водителей: %v", err)
		http.Error(w, "Не удалось загрузить список водителей", http.StatusInternalServerError)
		return
	}

	routes, err := h.service.GetRoutes(ctx)
	if err != nil {
		log.Printf("Ошибка получения маршрутов: %v", err)
		http.Error(w, "Не удалось загрузить список маршрутов", http.StatusInternalServerError)
		return
	}

	driversAutos := make(map[int][]models.Auto)
	for _, driver := range drivers {
		driverAutos, err := h.service.GetAutosByDriverID(ctx, driver.ID)
		if err != nil {
			log.Printf("Ошибка получения авто для водителя %d: %v", driver.ID, err)
			continue
		}
		driversAutos[driver.ID] = driverAutos
	}

	layoutBytes, err := os.ReadFile("./ui/template/layout.html")
	if err != nil {
		log.Printf("Ошибка чтения layout: %v", err)
		http.Error(w, "Ошибка чтения шаблона", http.StatusInternalServerError)
		return
	}

	editJournalBytes, err := os.ReadFile("./ui/template/journal_table/edit_journal.html")
	if err != nil {
		log.Printf("Ошибка чтения edit_journal: %v", err)
		http.Error(w, "Ошибка чтения шаблона", http.StatusInternalServerError)
		return
	}

	funcMap := template.FuncMap{
		"json": func(v interface{}) template.JS {
			a, _ := json.Marshal(v)
			return template.JS(a)
		},
	}

	tmpl := template.New("layout").Funcs(funcMap)

	tmpl, err = tmpl.Parse(string(layoutBytes))
	if err != nil {
		log.Printf("Ошибка парсинга layout: %v", err)
		http.Error(w, "Ошибка парсинга шаблона layout", http.StatusInternalServerError)
		return
	}

	tmpl, err = tmpl.Parse(string(editJournalBytes))
	if err != nil {
		log.Printf("Ошибка парсинга edit_journal: %v", err)
		http.Error(w, "Ошибка парсинга шаблона edit_journal", http.StatusInternalServerError)
		return
	}
	userName, err := h.getUserName(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	data := struct {
		Title        string
		Entry        *models.JournalView
		TimeOut      string // Добавляем это поле
		Drivers      []models.AutoPersonal
		DriversAutos map[int][]models.Auto
		Routes       []models.Route
		Username     string
	}{
		Title:        "Завершение рейса",
		Entry:        entry,
		TimeOut:      entry.TimeOut.Format("2006-01-02T15:04"),
		Drivers:      drivers,
		DriversAutos: driversAutos,
		Routes:       routes,
		Username:     userName,
	}

	err = tmpl.ExecuteTemplate(w, "layout", data)
	if err != nil {
		log.Printf("Ошибка выполнения шаблона: %v", err)
		http.Error(w, "Ошибка отображения страницы: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *AutoParkHandler) UpdateJournalEntry(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Ошибка парсинга формы", http.StatusBadRequest)
		return
	}

	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Некорректный ID записи журнала", http.StatusBadRequest)
		return
	}

	timeIn := r.Form.Get("time_in")
	if timeIn == "" {
		http.Error(w, "Время прибытия не может быть пустым", http.StatusBadRequest)
		return
	}

	err = h.service.CompleteJournalEntry(r.Context(), id, timeIn)
	if err != nil {
		log.Printf("Ошибка обновления записи журнала: %v", err)
		http.Error(w, "Не удалось обновить запись", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/journal", http.StatusSeeOther)
}

// Обработчик завершения рейса
func (h *AutoParkHandler) CompleteJournalEntry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	journalID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Некорректный ID записи журнала", http.StatusBadRequest)
		return
	}

	var requestBody struct {
		TimeIn string `json:"timeIn"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Некорректные данные", http.StatusBadRequest)
		return
	}

	if err := h.service.CompleteJournalEntry(r.Context(), journalID, requestBody.TimeIn); err != nil {
		http.Error(w, "Не удалось завершить рейс: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Рейс успешно завершен",
	})
}

// Удаление записи из журнала
func (h *AutoParkHandler) DeleteJournalEntry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	journalID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Некорректный ID записи журнала", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteJournalEntry(r.Context(), journalID)
	if err != nil {
		http.Error(w, "Не удалось удалить запись журнала: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Новый обработчик для страницы статистики
func (h *AutoParkHandler) StatisticsPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	routesVehicleCount, err := h.service.GetRoutesVehicleCount(ctx)
	if err != nil {
		log.Printf("Error getting routes vehicle count: %v", err)
		http.Error(w, "Не удалось получить статистику", http.StatusInternalServerError)
		return
	}
	userName, err := h.getUserName(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	data := struct {
		Title              string
		RoutesVehicleCount []models.RouteVehicleCount
		Username           string
	}{
		Title:              "Статистика маршрутов",
		RoutesVehicleCount: routesVehicleCount,
		Username:           userName,
	}

	tmpl, err := template.ParseFiles(
		"./ui/template/layout.html", "./ui/template/statistics.html",
	)
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Ошибка загрузки шаблона", http.StatusInternalServerError)
		return
	}

	w.Header().Del("Content-Type")

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
		return
	}
}
