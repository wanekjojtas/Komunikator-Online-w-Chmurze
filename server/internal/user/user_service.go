package user

import (
	"context"
	"fmt"
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

    emailExists, err := s.Repository.UserExistsByEmail(ctx, req.Email)
    if err != nil {
        return nil, err
    }
    if emailExists {
        return nil, fmt.Errorf("Email already exists")
    }

    usernameExists, err := s.Repository.UserExistsByUsername(ctx, req.Username)
    if err != nil {
        return nil, err
    }
    if usernameExists {
        return nil, fmt.Errorf("Username already exists")
    }

    hashedPassword, err := util.HashPassword(req.Password)
    if err != nil {
        return nil, err
    }

    u := &User{
        Username: req.Username,
        Email:    req.Email,
        Password: hashedPassword,
    }

    r, err := s.Repository.CreateUser(ctx, u)
    if err != nil {
        return nil, err
    }

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

	u, err := s.Repository.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return &LoginUserRes{}, err
	}

	err = util.CheckPassword(req.Password, u.Password)
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return nil, fmt.Errorf("Invalid password")
		}
		return &LoginUserRes{}, err // Preserve other unexpected errors
	}
	

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, util.MyJWTClaims{
		ID:       u.ID,
		Username: u.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    u.ID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	})

	ss, err := token.SignedString([]byte(config.GetSecretKey() ))
	if err != nil {
		return &LoginUserRes{}, err
	}

	return &LoginUserRes{accessToken: ss, Username: u.Username, ID: u.ID}, nil
}

func (s *service) SearchUsers(ctx context.Context, query string) ([]*User, error) {
    return s.Repository.SearchUsers(ctx, query)
}

