package user

import (
	"context"
	"database/sql"
	"fmt"
)

type repository struct {
	db DBTX
}

type DBTX interface {
	ExecContent(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

func NewRepository(db DBTX) Repository {
	return &repository{db: db}
}

// CreateUser inserts a new user into the database.
func (r *repository) CreateUser(ctx context.Context, user *User) (*User, error) {
	var id string
	query := "INSERT INTO users(username, password, email) VALUES ($1, $2, $3) RETURNING id"
	err := r.db.QueryRowContext(ctx, query, user.Username, user.Password, user.Email).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	user.ID = id
	return user, nil
}

// GetUserByEmail fetches a user by their email.
func (r *repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	u := User{}
	query := "SELECT id, email, username, password FROM users WHERE email = $1"
	err := r.db.QueryRowContext(ctx, query, email).Scan(&u.ID, &u.Email, &u.Username, &u.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Return nil for not found
		}
		return nil, fmt.Errorf("error fetching user by email: %w", err)
	}
	return &u, nil
}

// UserExistsByEmail checks if a user with the given email already exists.
func (r *repository) UserExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)"
	err := r.db.QueryRowContext(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error checking email existence: %w", err)
	}
	return exists, nil
}

// UserExistsByUsername checks if a user with the given username already exists.
func (r *repository) UserExistsByUsername(ctx context.Context, username string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)"
	err := r.db.QueryRowContext(ctx, query, username).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error checking username existence: %w", err)
	}
	return exists, nil
}

// SearchUsers performs a search for users by username.
func (r *repository) SearchUsers(ctx context.Context, query string) ([]*User, error) {
	var users []*User
	searchQuery := "%" + query + "%"

	rows, err := r.db.QueryContext(ctx, "SELECT id, username FROM users WHERE username ILIKE $1", searchQuery)
	if err != nil {
		return nil, fmt.Errorf("error querying users: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Username); err != nil {
			return nil, fmt.Errorf("error scanning user row: %w", err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over user rows: %w", err)
	}

	return users, nil
}