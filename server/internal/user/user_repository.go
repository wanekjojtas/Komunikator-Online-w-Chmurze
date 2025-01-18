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

func (r *repository) CreateUser(ctx context.Context, user *User) (*User, error) {
	var id string
	query := "INSERT INTO users(username, password, email) VALUES ($1, $2, $3) returning id"
	err := r.db.QueryRowContext(ctx, query, user.Username, user.Password, user.Email).Scan(&id)
	if err != nil {
		return &User{}, err
	}

	user.ID = id
	return user, nil
}

func (r *repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
    u := User{}
    query := "SELECT id, email, username, password FROM users WHERE email = $1"
    err := r.db.QueryRowContext(ctx, query, email).Scan(&u.ID, &u.Email, &u.Username, &u.Password)
    if err != nil {
        return nil, err
    }
    return &u, nil
}


func (r *repository) UserExistsByEmail(ctx context.Context, email string) (bool, error) {
    var exists bool
    query := "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)"
    err := r.db.QueryRowContext(ctx, query, email).Scan(&exists)
    return exists, err
}

func (r *repository) UserExistsByUsername(ctx context.Context, username string) (bool, error) {
    var exists bool
    query := "SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)"
    err := r.db.QueryRowContext(ctx, query, username).Scan(&exists)
    return exists, err
}

func (r *repository) SearchUsers(ctx context.Context, query string) ([]*User, error) {
    var users []*User
    query = "%" + query + "%"
    
    // Execute the query
    rows, err := r.db.QueryContext(ctx, "SELECT id, username FROM users WHERE username ILIKE $1", query)
    if err != nil {
        return nil, fmt.Errorf("error querying users: %w", err)
    }
    defer rows.Close()

    // Process each row
    for rows.Next() {
        var user User
        if err := rows.Scan(&user.ID, &user.Username); err != nil {
            return nil, fmt.Errorf("error scanning user row: %w", err)
        }
        users = append(users, &user)
    }

    // Handle rows iteration error
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating over user rows: %w", err)
    }

    return users, nil
}




