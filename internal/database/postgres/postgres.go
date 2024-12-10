package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"log"
	"time"

	"AutoParkWeb/internal/config"
	"AutoParkWeb/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresDB struct {
	Pool *pgxpool.Pool
}

func NewPostgresDB(cfg *config.Config) (*PostgresDB, error) {
	connString := cfg.GetPostgresConnectionString()
	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("unable to parse connection string: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %v", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("unable to ping database: %v", err)
	}

	log.Println("Successfully connected to PostgreSQL")

	return &PostgresDB{Pool: pool}, nil
}

func (db *PostgresDB) Close() {
	db.Pool.Close()
}

func (db *PostgresDB) withTransaction(ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, err := db.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback(ctx)

	if err := fn(tx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}
	return nil
}

// Методы для работы с водителями
func (db *PostgresDB) GetDrivers(ctx context.Context) ([]models.AutoPersonal, error) {
	var drivers []models.AutoPersonal

	query := `
		SELECT id, first_name, last_name, father_name
		FROM auto_personal
		ORDER BY first_name ASC
	`
	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error fetching drivers: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var driver models.AutoPersonal
		if err := rows.Scan(&driver.ID, &driver.FirstName, &driver.LastName, &driver.FatherName); err != nil {
			return nil, fmt.Errorf("error scanning driver row: %w", err)
		}
		drivers = append(drivers, driver)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return drivers, nil
}

func (db *PostgresDB) GetDriverByID(ctx context.Context, driverID int) (*models.AutoPersonal, error) {
	query := `
		SELECT id, first_name, last_name, father_name
		FROM auto_personal
		WHERE id = $1
	`
	row := db.Pool.QueryRow(ctx, query, driverID)

	var driver models.AutoPersonal
	err := row.Scan(&driver.ID, &driver.FirstName, &driver.LastName, &driver.FatherName)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New("driver not found")
		}
		return nil, err
	}

	return &driver, nil
}

func (db *PostgresDB) AddDriver(ctx context.Context, firstName, lastName, fatherName string) error {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		err := tx.Rollback(ctx)
		if err != nil {

		}
	}(tx, ctx)

	query := `INSERT INTO auto_personal (first_name, last_name, father_name) VALUES ($1, $2, $3)`
	_, err = tx.Exec(ctx, query, firstName, lastName, fatherName)
	if err != nil {
		return fmt.Errorf("failed to add driver: %v", err)
	}

	return tx.Commit(ctx)
}

func (db *PostgresDB) UpdateDriver(ctx context.Context, driverID int, firstName, lastName, fatherName string) error {
	return db.withTransaction(ctx, func(tx pgx.Tx) error {
		query := `UPDATE auto_personal SET first_name = $1, last_name = $2, father_name = $3 WHERE id = $4`
		_, err := tx.Exec(ctx, query, firstName, lastName, fatherName, driverID)
		if err != nil {
			return fmt.Errorf("failed to update driver: %v", err)
		}
		return nil
	})
}

func (db *PostgresDB) DeleteDriver(ctx context.Context, driverID int) error {
	return db.withTransaction(ctx, func(tx pgx.Tx) error {
		deleteCarsQuery := `DELETE FROM auto WHERE personal_id = $1`
		_, err := tx.Exec(ctx, deleteCarsQuery, driverID)
		if err != nil {
			return fmt.Errorf("failed to delete cars: %v", err)
		}

		deleteDriverQuery := `DELETE FROM auto_personal WHERE id = $1`
		_, err = tx.Exec(ctx, deleteDriverQuery, driverID)
		if err != nil {
			return fmt.Errorf("failed to delete driver: %v", err)
		}
		return nil
	})
}

// Методы для работы с автомобилями
func (db *PostgresDB) GetCars(ctx context.Context) ([]models.Auto, error) {
	var cars []models.Auto

	query := `
		SELECT a.id, a.num, a.color, a.mark, a.personal_id, 
		       CONCAT(p.last_name, ' ', p.first_name, ' ', p.father_name) AS driver_name
		FROM auto a
		LEFT JOIN auto_personal p ON a.personal_id = p.id
		ORDER BY a.num ASC
	`

	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error fetching cars: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var car models.Auto
		var driverName string

		if err := rows.Scan(&car.ID, &car.Num, &car.Color, &car.Mark, &car.PersonalID, &driverName); err != nil {
			return nil, fmt.Errorf("error scanning car row: %w", err)
		}

		car.DriverFullName = driverName

		cars = append(cars, car)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return cars, nil
}

func (db *PostgresDB) GetCarByID(ctx context.Context, carID int) (*models.Auto, string, error) {
	query := `
		SELECT a.id, a.num, a.color, a.mark, a.personal_id, 
		       CONCAT(p.last_name, ' ', p.first_name, ' ', p.father_name) AS driver_full_name
		FROM auto a
		LEFT JOIN auto_personal p ON a.personal_id = p.id
		WHERE a.id = $1
	`
	row := db.Pool.QueryRow(ctx, query, carID)

	var car models.Auto
	var driverFullName sql.NullString
	err := row.Scan(&car.ID, &car.Num, &car.Color, &car.Mark, &car.PersonalID, &driverFullName)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, "", errors.New("car not found")
		}
		return nil, "", err
	}

	if driverFullName.Valid {
		return &car, driverFullName.String, nil
	}

	return &car, "", nil
}

func (db *PostgresDB) AddCar(ctx context.Context, num, color, mark string, personalID int) error {
	return db.withTransaction(ctx, func(tx pgx.Tx) error {
		query := `INSERT INTO auto (num, color, mark, personal_id) VALUES ($1, $2, $3, $4)`
		_, err := tx.Exec(ctx, query, num, color, mark, personalID)
		if err != nil {
			return fmt.Errorf("failed to add car: %v", err)
		}
		return nil
	})
}

func (db *PostgresDB) UpdateCar(ctx context.Context, carID int, num, color, mark string, personalID int) error {
	return db.withTransaction(ctx, func(tx pgx.Tx) error {
		query := `UPDATE auto SET num = $1, color = $2, mark = $3, personal_id = $4 WHERE id = $5`
		_, err := tx.Exec(ctx, query, num, color, mark, personalID, carID)
		if err != nil {
			return fmt.Errorf("не удалось обновить автомобиль с ID %d: %v", carID, err)
		}
		return nil
	})
}

func (db *PostgresDB) DeleteCar(ctx context.Context, carID int) error {
	var exists bool
	checkQuery := `SELECT EXISTS (SELECT 1 FROM auto WHERE id = $1)`
	if err := db.Pool.QueryRow(ctx, checkQuery, carID).Scan(&exists); err != nil {
		return fmt.Errorf("failed to check if car exists: %v", err)
	}

	if !exists {
		return sql.ErrNoRows
	}

	var hasJournalRecords bool
	journalQuery := `SELECT EXISTS (SELECT 1 FROM journal WHERE auto_id = $1)`
	if err := db.Pool.QueryRow(ctx, journalQuery, carID).Scan(&hasJournalRecords); err != nil {
		return fmt.Errorf("failed to check journal_table records: %v", err)
	}

	return db.withTransaction(ctx, func(tx pgx.Tx) error {
		log.Printf("Attempting to delete car with ID: %d", carID)

		if hasJournalRecords {
			deleteJournalQuery := `DELETE FROM journal WHERE auto_id = $1`
			_, err := tx.Exec(ctx, deleteJournalQuery, carID)
			if err != nil {
				return fmt.Errorf("failed to delete car journal_table records: %v", err)
			}
		}

		deleteCarQuery := `DELETE FROM auto WHERE id = $1`
		result, err := tx.Exec(ctx, deleteCarQuery, carID)
		if err != nil {
			return fmt.Errorf("failed to delete car: %v", err)
		}

		rowsAffected := result.RowsAffected()
		if rowsAffected == 0 {
			return fmt.Errorf("no car deleted, possible concurrent modification")
		}

		return nil
	})
}

// Методы для работы с маршрутами
func (db *PostgresDB) GetRoutes(ctx context.Context) ([]models.Route, error) {
	var routes []models.Route

	query := "SELECT id, start_point, end_point FROM routes ORDER BY id ASC"
	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error fetching routes: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var route models.Route
		if err := rows.Scan(&route.ID, &route.StartPoint, &route.EndPoint); err != nil {
			return nil, fmt.Errorf("error scanning route row: %w", err)
		}
		routes = append(routes, route)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return routes, nil
}

func (db *PostgresDB) GetRouteByID(ctx context.Context, routeID int) (*models.Route, error) {
	query := `SELECT id, start_point, end_point FROM routes WHERE id = $1`
	row := db.Pool.QueryRow(ctx, query, routeID)

	var route models.Route
	err := row.Scan(&route.ID, &route.StartPoint, &route.EndPoint)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New("route not found")
		}
		return nil, err
	}

	return &route, nil
}

func (db *PostgresDB) AddRoute(ctx context.Context, startPoint, endPoint string) error {
	return db.withTransaction(ctx, func(tx pgx.Tx) error {
		query := `INSERT INTO routes (start_point, end_point) VALUES ($1, $2)`
		_, err := tx.Exec(ctx, query, startPoint, endPoint)
		if err != nil {
			return fmt.Errorf("failed to add route: %v", err)
		}
		return nil
	})
}

func (db *PostgresDB) UpdateRoute(ctx context.Context, route *models.Route) error {
	return db.withTransaction(ctx, func(tx pgx.Tx) error {
		query := `UPDATE routes SET start_point = $1, end_point = $2 WHERE id = $3`
		_, err := tx.Exec(ctx, query, route.StartPoint, route.EndPoint, route.ID)
		if err != nil {
			return fmt.Errorf("failed to update route: %v", err)
		}
		return nil
	})
}

func (db *PostgresDB) DeleteRoute(ctx context.Context, routeID int) error {
	return db.withTransaction(ctx, func(tx pgx.Tx) error {
		query := `DELETE FROM routes WHERE id = $1`
		_, err := tx.Exec(ctx, query, routeID)
		if err != nil {
			return fmt.Errorf("failed to delete route: %v", err)
		}
		return nil
	})
}

// Методы для работы с журналом
func (db *PostgresDB) GetAllJournalEntries(ctx context.Context) ([]models.JournalView, error) {
	var entries []models.JournalView
	query := "SELECT * FROM journal_view ORDER BY journal_view.time_out"
	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all journal_table entries: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var entry models.JournalView
		if err := rows.Scan(&entry.JournalID, &entry.TimeOut, &entry.TimeIn, &entry.StartPoint, &entry.EndPoint, &entry.AutoNumber, &entry.AutoMark, &entry.DriverName); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func (db *PostgresDB) GetJournalEntryByID(ctx context.Context, journalID int) (*models.JournalView, error) {
	var entry models.JournalView
	query := "SELECT * FROM journal_view WHERE journal_id = $1"
	err := db.Pool.QueryRow(ctx, query, journalID).Scan(&entry.JournalID, &entry.TimeOut, &entry.TimeIn, &entry.StartPoint, &entry.EndPoint, &entry.AutoNumber, &entry.AutoMark, &entry.DriverName)
	if err != nil {
		return nil, fmt.Errorf("failed to get journal_table entry by ID: %v", err)
	}
	return &entry, nil
}

func (db *PostgresDB) GetAutosByDriverID(ctx context.Context, driverID int) ([]models.Auto, error) {
	query := `SELECT id, num, color, mark, personal_id FROM auto WHERE personal_id = $1`

	rows, err := db.Pool.Query(ctx, query, driverID)
	if err != nil {
		return nil, fmt.Errorf("failed to query autos for driver %d: %w", driverID, err)
	}
	defer rows.Close()

	var autos []models.Auto
	for rows.Next() {
		var auto models.Auto
		err := rows.Scan(
			&auto.ID,
			&auto.Num,
			&auto.Color,
			&auto.Mark,
			&auto.PersonalID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan auto row for driver %d: %w", driverID, err)
		}
		autos = append(autos, auto)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during auto rows iteration for driver %d: %w", driverID, err)
	}

	return autos, nil
}

func (db *PostgresDB) AddJournalEntry(ctx context.Context, autoID, routeID int, timeOut time.Time) error {
	return db.withTransaction(ctx, func(tx pgx.Tx) error { // Используем pgx.Tx
		query := `INSERT INTO journal (auto_id, route_id, time_out) VALUES ($1, $2, $3)`
		_, err := tx.Exec(ctx, query, autoID, routeID, timeOut)
		if err != nil {
			return fmt.Errorf("failed to add journal_table entry: %v", err)
		}
		return nil
	})
}

func (db *PostgresDB) CompleteJournalEntry(ctx context.Context, entryID int, timeIn time.Time) error {
	return db.withTransaction(ctx, func(tx pgx.Tx) error {
		query := `UPDATE journal SET time_in = $1 WHERE id = $2`
		_, err := tx.Exec(ctx, query, timeIn, entryID)
		if err != nil {
			return fmt.Errorf("failed to update journal_table entry: %v", err)
		}
		return nil
	})
}

func (db *PostgresDB) DeleteJournalEntry(ctx context.Context, entryID int) error {
	return db.withTransaction(ctx, func(tx pgx.Tx) error { // Используем pgx.Tx
		query := `DELETE FROM journal WHERE id = $1`
		_, err := tx.Exec(ctx, query, entryID)
		if err != nil {
			return fmt.Errorf("failed to delete journal_table entry: %v", err)
		}
		return nil
	})
}

// Получение количества машин на каждом маршруте
func (db *PostgresDB) GetRoutesVehicleCount(ctx context.Context) ([]models.RouteVehicleCount, error) {
	query := `SELECT * FROM get_routes_vehicle_count()`
	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %v", err)
	}
	defer rows.Close()

	var routes []models.RouteVehicleCount
	for rows.Next() {
		var route models.RouteVehicleCount
		if err := rows.Scan(&route.RouteName, &route.VehicleCount); err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
		routes = append(routes, route)
	}

	return routes, nil
}
