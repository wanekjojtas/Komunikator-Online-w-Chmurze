package user

import (
	"context"
	"fmt"
	"log"
	"server/config"
	"server/util"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

type service struct {
	Repository
	timeout time.Duration
}

func NewService(repository Repository) Service {
	return &service{
		repository,
		time.Duration(2) * time.Second,
	}
}

func (s *service) CreateUser(c context.Context, req *CreateUserReq) (*CreateUserRes, error) {
    ctx, cancel := context.WithTimeout(c, s.timeout)
    defer cancel()

    log.Printf("Creating user with email: %s", req.Email)

    // Validate email format
    if !isValidEmail(req.Email) {
        log.Printf("Invalid email format: %s", req.Email)
        return nil, fmt.Errorf("invalid_email_format")
    }

    // Validate password length
    if len(req.Password) < 8 {
        log.Printf("Password too short for email: %s", req.Email)
        return nil, fmt.Errorf("password_too_short")
    }

    // Check if email already exists
    emailExists, err := s.Repository.UserExistsByEmail(ctx, req.Email)
    if err != nil {
        log.Printf("Error checking email existence: %v", err)
        return nil, fmt.Errorf("internal_error")
    }
    if emailExists {
        log.Printf("Email already exists: %s", req.Email)
        return nil, fmt.Errorf("email_already_exists")
    }

    // Check if username already exists
    usernameExists, err := s.Repository.UserExistsByUsername(ctx, req.Username)
    if err != nil {
        log.Printf("Error checking username existence: %v", err)
        return nil, fmt.Errorf("internal_error")
    }
    if usernameExists {
        log.Printf("Username already exists: %s", req.Username)
        return nil, fmt.Errorf("username_already_exists")
    }

    // Hash the password
    hashedPassword, err := util.HashPassword(req.Password)
    if err != nil {
        log.Printf("Error hashing password for email: %s, error: %v", req.Email, err)
        return nil, fmt.Errorf("internal_error")
    }

    // Create user entity
    u := &User{
        Username: req.Username,
        Email:    req.Email,
        Password: hashedPassword,
    }

    // Save user to the repository
    r, err := s.Repository.CreateUser(ctx, u)
    if err != nil {
        log.Printf("Error creating user in repository for email: %s, error: %v", req.Email, err)
        return nil, fmt.Errorf("internal_error")
    }

    log.Printf("User created successfully: ID=%s, Username=%s", r.ID, r.Username)

    // Prepare response
    res := &CreateUserRes{
        ID:       r.ID,
        Username: r.Username,
        Email:    r.Email,
    }

    return res, nil
}

func (s *service) Login(c context.Context, req *LoginUserReq) (*LoginUserRes, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	log.Printf("Login attempt for email: %s", req.Email)

	if !isValidEmail(req.Email) {
		log.Printf("Invalid email format: %s", req.Email)
		return nil, fmt.Errorf("Invalid email format")
	}

	u, err := s.Repository.GetUserByEmail(ctx, req.Email)
	if err != nil {
		log.Printf("Error fetching user by email: %v", err)
		return nil, fmt.Errorf("User not found")
	}

	log.Printf("User found: ID=%s, Username=%s", u.ID, u.Username)

	err = util.CheckPassword(req.Password, u.Password)
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			log.Printf("Password mismatch for user ID=%s", u.ID)
			return nil, fmt.Errorf("Invalid password")
		}
		log.Printf("Error checking password: %v", err)
		return nil, err
	}

	log.Printf("Password validated successfully for user ID=%s", u.ID)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, util.MyJWTClaims{
		ID:       u.ID,
		Username: u.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    u.ID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	})

	ss, err := token.SignedString([]byte(config.GetSecretKey()))
	if err != nil {
		log.Printf("Error signing token for user ID=%s: %v", u.ID, err)
		return nil, err
	}

	log.Printf("Token generated successfully for user ID=%s", u.ID)

	return &LoginUserRes{accessToken: ss, Username: u.Username, ID: u.ID}, nil
}

func (s *service) SearchUsers(ctx context.Context, query string) ([]*User, error) {
	log.Printf("Searching users with query: %s", query)
	users, err := s.Repository.SearchUsers(ctx, query)
	if err != nil {
		log.Printf("Error searching users: %v", err)
		return nil, err
	}

	log.Printf("SearchUsers completed successfully: %d users found", len(users))
	return users, nil
}
