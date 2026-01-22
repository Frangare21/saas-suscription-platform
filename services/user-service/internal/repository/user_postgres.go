package repository

import (
	"context"
	"errors"
	"saas-subscription-platform/services/user-service/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrUserNotFound = errors.New("user not found")
var ErrUserExists = errors.New("user already exists")

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(email, password string) (model.User, error) {
	user := model.User{
		ID:       generateUUID(),
		Email:    email,
		Password: password,
	}

	query := `
		INSERT INTO users (id, email, password)
		VALUES ($1, $2, $3)
		RETURNING created_at
	`

	err := r.db.QueryRow(
		context.Background(),
		query,
		user.ID,
		user.Email,
		user.Password,
	).Scan(&user.CreatedAt)

	if err != nil {
		if isUniqueViolation(err) {
			return model.User{}, ErrUserExists
		}
		return model.User{}, err
	}

	return user, nil
}

func (r *UserRepository) GetByEmail(email string) (model.User, error) {
	var user model.User

	query := `
		SELECT id, email, password, created_at
		FROM users
		WHERE email = $1
	`

	err := r.db.QueryRow(context.Background(), query, email).
		Scan(&user.ID, &user.Email, &user.Password, &user.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrUserNotFound
		}
		return model.User{}, err
	}

	return user, nil
}

func (r *UserRepository) GetByID(userID string) (model.User, error) {
	var user model.User

	query := `
		SELECT id, email, password, created_at
		FROM users
		WHERE id = $1
	`

	err := r.db.QueryRow(context.Background(), query, userID).
		Scan(&user.ID, &user.Email, &user.Password, &user.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrUserNotFound
		}
		return model.User{}, err
	}

	return user, nil
}

// UpdateFields actualiza email y/o password (si el puntero es nil, mantiene el valor actual).
func (r *UserRepository) UpdateFields(userID string, email *string, password *string) error {
	query := `
		UPDATE users
		SET
			email = COALESCE($1, email),
			password = COALESCE($2, password)
		WHERE id = $3
	`

	ct, err := r.db.Exec(context.Background(), query, email, password, userID)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrUserExists
		}
		return err
	}

	if ct.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *UserRepository) Delete(userID string) error {
	query := `DELETE FROM users WHERE id = $1`
	ct, err := r.db.Exec(context.Background(), query, userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}
