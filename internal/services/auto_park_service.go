package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"AutoParkWeb/internal/auth"
	"AutoParkWeb/internal/database/postgres"
	"AutoParkWeb/internal/models"
)

type AutoParkService struct {
	db database.DBHandler
}

func NewAutoParkService(db *database.PostgresDB) *AutoParkService {
	return &AutoParkService{db: db}
}

// Методы для работы с водителями
func (s *AutoParkService) GetDrivers(ctx context.Context) ([]models.AutoPersonal, error) {
	return s.db.GetDrivers(ctx)
}

func (s *AutoParkService) GetDriverByID(ctx context.Context, driverID int) (*models.AutoPersonal, error) {
	return s.db.GetDriverByID(ctx, driverID)
}

func (s *AutoParkService) AddDriver(ctx context.Context, firstName, lastName, fatherName string) error {
	if firstName == "" || lastName == "" {
		return errors.New("first name and last name are required")
	}
	return s.db.AddDriver(ctx, firstName, lastName, fatherName)
}

func (s *AutoParkService) UpdateDriver(ctx context.Context, driverID int, firstName, lastName, fatherName string) error {
	if firstName == "" || lastName == "" {
		return errors.New("first name and last name are required")
	}
	return s.db.UpdateDriver(ctx, driverID, firstName, lastName, fatherName)
}

func (s *AutoParkService) DeleteDriver(ctx context.Context, driverID int) error {
	return s.db.DeleteDriver(ctx, driverID)
}

// Методы для работы с автомобилями

func (s *AutoParkService) GetCars(ctx context.Context) ([]models.Auto, error) {
	return s.db.GetCars(ctx)
}

func (s *AutoParkService) GetCarByID(ctx context.Context, carID int) (*models.Auto, string, error) {
	return s.db.GetCarByID(ctx, carID)
}

func (s *AutoParkService) AddCar(ctx context.Context, num, color, mark string, personalID int) error {
	if num == "" || color == "" || mark == "" {
		return errors.New("num, color and mark are required")
	}
	err := s.db.AddCar(ctx, num, color, mark, personalID)
	if err != nil {
		return fmt.Errorf("не удалось добавить автомобиль: %v", err)
	}
	return nil
}

func (s *AutoParkService) UpdateCar(ctx context.Context, carID int, num, color, mark string, personalID int) error {
	if num == "" || color == "" || mark == "" {
		return errors.New("num, color and mark are required")
	}

	err := s.db.UpdateCar(ctx, carID, num, color, mark, personalID)
	if err != nil {
		return fmt.Errorf("не удалось обновить автомобиль с ID %d: %v", carID, err)
	}

	return nil
}

func (s *AutoParkService) DeleteCar(ctx context.Context, carID int) error {
	return s.db.DeleteCar(ctx, carID)
}

// Методы для работы с маршрутами
func (s *AutoParkService) GetRoutes(ctx context.Context) ([]models.Route, error) {
	return s.db.GetRoutes(ctx)
}

func (s *AutoParkService) GetRouteByID(ctx context.Context, routeID int) (*models.Route, error) {
	return s.db.GetRouteByID(ctx, routeID)
}

func (s *AutoParkService) AddRoute(ctx context.Context, startPoint, endPoint string) error {
	if startPoint == "" || endPoint == "" {
		return errors.New("both startPoint and endPoint are required")
	}
	return s.db.AddRoute(ctx, startPoint, endPoint)
}

func (s *AutoParkService) UpdateRoute(ctx context.Context, route *models.Route) error {
	return s.db.UpdateRoute(ctx, route)
}

func (s *AutoParkService) DeleteRoute(ctx context.Context, routeID int) error {
	return s.db.DeleteRoute(ctx, routeID)
}

// Методы для работы с журналом
func (s *AutoParkService) GetAllJournalEntries(ctx context.Context) ([]models.JournalView, error) {
	entries, err := s.db.GetAllJournalEntries(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all journal_table entries: %v", err)
	}
	return entries, nil
}

func (s *AutoParkService) GetJournalEntryByID(ctx context.Context, journalID int) (*models.JournalView, error) {
	if journalID <= 0 {
		return nil, errors.New("invalid journal_table ID")
	}
	entry, err := s.db.GetJournalEntryByID(ctx, journalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get journal_table entry by ID: %v", err)
	}
	return entry, nil
}

func (s *AutoParkService) GetAutosByDriverID(ctx context.Context, driverID int) ([]models.Auto, error) {
	autos, err := s.db.GetAutosByDriverID(ctx, driverID)
	if err != nil {
		return nil, fmt.Errorf("error getting autos for driver with ID %d: %v", driverID, err)
	}

	return autos, nil
}

func (s *AutoParkService) AddJournalEntry(ctx context.Context, autoID, routeID int, timeOut string) error {
	if autoID <= 0 || routeID <= 0 {
		return errors.New("autoID and routeID must be positive")
	}
	if timeOut == "" {
		return errors.New("time out is required")
	}

	timeOutParsed, err := time.Parse("2006-01-02T15:04", timeOut)
	if err != nil {
		return fmt.Errorf("invalid timeOut format: %v", err)
	}

	return s.db.AddJournalEntry(ctx, autoID, routeID, timeOutParsed)
}

func (s *AutoParkService) CompleteJournalEntry(ctx context.Context, entryID int, timeIn string) error {
	if entryID <= 0 {
		return errors.New("entryID must be positive")
	}
	if timeIn == "" {
		return errors.New("time in is required")
	}

	timeInParsed, err := time.Parse("2006-01-02T15:04", timeIn)
	if err != nil {
		return fmt.Errorf("invalid timeIn format: %v", err)
	}

	return s.db.CompleteJournalEntry(ctx, entryID, timeInParsed)
}

func (s *AutoParkService) DeleteJournalEntry(ctx context.Context, entryID int) error {
	if entryID <= 0 {
		return errors.New("invalid entry ID")
	}
	return s.db.DeleteJournalEntry(ctx, entryID)
}

// Методы для аналитики
func (s *AutoParkService) GetRoutesVehicleCount(ctx context.Context) ([]models.RouteVehicleCount, error) {
	return s.db.GetRoutesVehicleCount(ctx)
}

// Методы для работы с пользователями
func (s *AutoParkService) AuthenticateUser(username, password string) (*models.User, error) {
	user, err := s.db.GetUserByUsername(context.Background(), username)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	if !auth.CheckPasswordHash(password, user.PasswordHash) {
		return nil, fmt.Errorf("invalid password")
	}

	return user, nil
}

func (s *AutoParkService) RegisterUser(username, password string) error {
	username = strings.TrimSpace(username)
	if len(username) < 3 {
		return fmt.Errorf("имя пользователя должно содержать не менее 3 символов")
	}

	if len(password) < 6 {
		return fmt.Errorf("пароль должен содержать не менее 6 символов")
	}

	_, err := s.db.GetUserByUsername(context.Background(), username)
	if err == nil {
		return fmt.Errorf("пользователь с таким именем уже существует")
	}

	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		return fmt.Errorf("ошибка хэширования пароля: %v", err)
	}

	err = s.db.AddUser(context.Background(), username, hashedPassword, "user")
	if err != nil {
		return fmt.Errorf("ошибка создания пользователя: %v", err)
	}

	return nil
}
