package repository

import (
	"context"

	"github.com/ciaranmcdonnell/go-api-server/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserQueriesInterface interface {
	CreateUser(ctx context.Context, user models.User) (int64, error)
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
	GetUserByID(ctx context.Context, id int64) (models.User, error)
}

type UserQueries struct {
	db *pgxpool.Pool
}

func NewUserQueries(db *pgxpool.Pool) UserQueriesInterface {
	return &UserQueries{db: db}
}

func (q *UserQueries) CreateUser(ctx context.Context, user models.User) (int64, error) {
	var id int64

	query := `INSERT INTO users (email, name, password_hash)
	VALUES ($1, $2, $3)
	RETURNING id`

	err := q.db.QueryRow(ctx, query, user.Email, user.Name, user.HashedPassword).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (q *UserQueries) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	var user models.User

	query := `SELECT id, email, name, password_hash FROM users
	WHERE email = $1
	LIMIT 1`

	err := q.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Password,
	)
	if err != nil {
		return user, err
	}
	return user, nil
}

func (q *UserQueries) GetUserByID(ctx context.Context, id int64) (models.User, error) {
	var user models.User

	query := `SELECT id, email, name, role, created_at FROM users
	WHERE id = $1
	LIMIT 1`

	err := q.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Role,
		&user.CreatedAt,
	)
	if err != nil {
		return user, err
	}
	return user, nil
}
