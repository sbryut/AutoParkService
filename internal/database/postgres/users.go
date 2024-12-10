package database

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"

	"AutoParkWeb/internal/models"
)

// Метод для добавления нового пользователя в базу данных
func (db *PostgresDB) AddUser(ctx context.Context, username, passwordHash, role string) error {
	query := `INSERT INTO users (username, password_hash, role) VALUES ($1, $2, $3)`
	_, err := db.Pool.Exec(ctx, query, username, passwordHash, role)
	if err != nil {
		return fmt.Errorf("failed to add user: %v", err)
	}
	return nil
}

// Метод для получения пользователя по имени
func (db *PostgresDB) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `SELECT id, username, password_hash, role FROM users WHERE username = $1`
	row := db.Pool.QueryRow(ctx, query, username)

	var user models.User
	err := row.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	return &user, nil
}
