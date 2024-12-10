package database

import (
	"context"
	"time"

	"AutoParkWeb/internal/models"
)

type DBHandler interface {
	// Методы для работы с водителями
	GetDrivers(ctx context.Context) ([]models.AutoPersonal, error)
	GetDriverByID(ctx context.Context, driverID int) (*models.AutoPersonal, error)
	AddDriver(ctx context.Context, firstName, lastName, fatherName string) error
	UpdateDriver(ctx context.Context, driverID int, firstName, lastName, fatherName string) error
	DeleteDriver(ctx context.Context, driverID int) error

	// Методы для работы с автомобилями
	GetCars(ctx context.Context) ([]models.Auto, error)
	GetCarByID(ctx context.Context, carID int) (*models.Auto, string, error)
	AddCar(ctx context.Context, num, color, mark string, personalID int) error
	UpdateCar(ctx context.Context, carID int, num, color, mark string, personalID int) error
	DeleteCar(ctx context.Context, carID int) error

	// Методы для работы с маршрутами
	GetRoutes(ctx context.Context) ([]models.Route, error)
	GetRouteByID(ctx context.Context, routeID int) (*models.Route, error)
	AddRoute(ctx context.Context, startPoint, endPoint string) error
	UpdateRoute(ctx context.Context, route *models.Route) error
	DeleteRoute(ctx context.Context, routeID int) error

	// Методы для работы с журналом
	GetAllJournalEntries(ctx context.Context) ([]models.JournalView, error)
	GetJournalEntryByID(ctx context.Context, journalID int) (*models.JournalView, error)
	GetAutosByDriverID(ctx context.Context, driverID int) ([]models.Auto, error)
	AddJournalEntry(ctx context.Context, autoID, routeID int, timeOut time.Time) error
	CompleteJournalEntry(ctx context.Context, entryID int, timeIn time.Time) error
	DeleteJournalEntry(ctx context.Context, entryID int) error

	// Процедуры для аналитики
	GetRoutesVehicleCount(ctx context.Context) ([]models.RouteVehicleCount, error)

	// Методы для работы с пользователями автопарка
	AddUser(ctx context.Context, username, passwordHash, role string) error
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
}
