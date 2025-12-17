package userdb

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const UniqueViolation = "23505"

var (
	ErrUsernameAlreadyExists = errors.New("username already exists")
	ErrEmailAlreadyExists    = errors.New("email already exists")
	ErrRecordNotFound        = errors.New("record not found")
)

type Store struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) Create(user *User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		INSERT INTO 
			users (id, username, email, password_hash, activated)
		VALUES
			($1, $2, $3, $4, $5)
		RETURNING 
			created_at, version
	`

	args := []any{user.ID, user.Username, user.Email, user.Password.hash, user.Activated}

	err := s.pool.QueryRow(ctx, query, args...).Scan(&user.CreatedAt, &user.Version)
	if err != nil {
		var pgErr *pgconn.PgError
		switch {
		case errors.As(err, &pgErr) && pgErr.Code == UniqueViolation:
			switch pgErr.ConstraintName {
			case "users_username_key":
				return ErrUsernameAlreadyExists
			case "users_email_key":
				return ErrEmailAlreadyExists
			}
		default:
			return err
		}
	}

	return nil
}

func (s *Store) UpdateUser(user *User) error {
	query := `
		UPDATE 
			users
		SET 
			username 		= $1, 
			email 			= $2, 
			password_hash 	= $3, 
			last_login 		= $4, 
			activated 		= $5, 
			version 		= version + 1
		WHERE	
			id = $6 
		AND 
			version = $7
		RETURNING
			version
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{
		user.Username,
		user.Email,
		user.Password.hash,
		user.LastLogin,
		user.Activated,
		user.ID,
		user.Version,
	}

	err := s.pool.QueryRow(ctx, query, args...).Scan(&user.Version)
	if err != nil {
		var pgErr *pgconn.PgError
		switch {
		case errors.As(err, &pgErr) && pgErr.Code == UniqueViolation:
			switch pgErr.ConstraintName {
			case "users_username_key":
				return ErrUsernameAlreadyExists
			case "users_email_key":
				return ErrEmailAlreadyExists
			}
		default:
			return err
		}
	}

	return nil
}

func (s *Store) DeleteUser(username string) error {
	query := `
		DELETE FROM 
			users
		WHERE 
			username = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmdTag, err := s.pool.Exec(ctx, query, username)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrRecordNotFound
	}

	return nil
}

func (s *Store) GetUserByEmail(email string) (*User, error) {
	query := `
		SELECT 
			id, username, email, password_hash, 
			created_at, last_login, activated, version
		FROM
			users
		WHERE
			email = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var user User

	err := s.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password.hash,
		&user.CreatedAt,
		&user.LastLogin,
		&user.Activated,
		&user.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (s *Store) GetUserByUsername(username string) (*User, error) {
	query := `
		SELECT 
			id, username, email, password_hash, 
			created_at, last_login, activated, version
		FROM
			users
		WHERE
			username = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var user User

	err := s.pool.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password.hash,
		&user.CreatedAt,
		&user.LastLogin,
		&user.Activated,
		&user.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (s *Store) GetUsers() ([]*User, error) {
	query := `
		SELECT 
			id, username, email, password_hash, 
			created_at, last_login, activated, version
		FROM
			users
		ORDER BY 
			username ASC
		LIMIT 
			10	
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]*User, 0)

	for rows.Next() {
		var user User

		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Password.hash,
			&user.CreatedAt,
			&user.LastLogin,
			&user.Activated,
			&user.Version,
		)

		if err != nil {
			return nil, err
		}

		users = append(users, &user)
	}

	if rows.Err() != nil {
		return nil, err
	}

	return users, nil
}
