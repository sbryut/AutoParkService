package models

import "time"

type AutoPersonal struct {
	ID         int    `db:"id"`
	FirstName  string `db:"first_name"`
	LastName   string `db:"last_name"`
	FatherName string `db:"father_name"`
}

type Auto struct {
	ID             int    `db:"id"`
	Num            string `db:"num"`
	Color          string `db:"color"`
	Mark           string `db:"mark"`
	PersonalID     int    `db:"personal_id"`
	DriverFullName string `db:"driver_full_name"`
}

type Route struct {
	ID         int     `db:"id"`
	StartPoint string  `db:"start_point"`
	EndPoint   string  `db:"end_point"`
	Name       string  `db:"name"`
	TimeDiff   float64 `json:"time_diff"`
}

type JournalView struct {
	JournalID  int        `db:"journal_id"`
	TimeOut    time.Time  `db:"time_out"`
	TimeIn     *time.Time `db:"time_in"`
	StartPoint string     `db:"start_point"`
	EndPoint   string     `db:"end_point"`
	AutoNumber string     `db:"auto_number"`
	AutoMark   string     `db:"auto_mark"`
	DriverName string     `db:"driver_name"`
}

type User struct {
	ID           int
	Username     string
	PasswordHash string
	Role         string
	CreatedAt    time.Time `json:"created_at"`
}

type RouteTime struct {
	RouteName   string  `json:"route_name"`
	AverageTime float64 `json:"average_time"`
}

type RouteVehicleCount struct {
	RouteName    string `json:"route_name"`
	VehicleCount int64  `json:"vehicle_count"`
}
