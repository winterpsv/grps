package service

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"strings"
	repository "task4_1/user-management/internal/adapter/db/mongodb"
	"task4_1/user-management/internal/infrastructure/config"
	"task4_1/user-management/pkg/apperror"
	"time"
)

type Auth struct {
	UserRepository repository.UserRepositoryInterface
	cfg            *config.Config
}

func NewAuth(r repository.UserRepositoryInterface, cfg *config.Config) *Auth {
	return &Auth{r, cfg}
}

func (au *Auth) GenerateHash(password string) string {
	h := sha256.New()

	h.Write([]byte(password))

	return hex.EncodeToString(h.Sum(nil))
}

func (au *Auth) comparePasswordHash(nickname, password string) bool {
	passwordHash := au.GenerateHash(password)

	u, err := au.UserRepository.FindByNickname(nickname)
	if err != nil {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(u.PasswordHash), []byte(passwordHash)) == 1
}

func (au *Auth) GenerateToken(nickname, password string) (string, error) {
	if !au.comparePasswordHash(nickname, password) {
		return "", apperror.ErrAccess
	}

	user, err := au.UserRepository.FindByNickname(nickname)
	if err != nil {
		return "", apperror.ErrFindUser
	}

	// Generate JWT token
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = user.ID
	claims["role"] = user.Role
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	// Sign and encode JWT token
	jwtSecret := []byte(au.cfg.JwtSecret)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", apperror.ErrCreateToken
	}

	return tokenString, nil
}

func (au *Auth) parseHeader(req *http.Request) (string, error) {
	authHeader := req.Header.Get("Authorization")
	if authHeader == "" {
		return "", apperror.ErrAuthorizationToken
	}

	auth := strings.SplitN(authHeader, " ", 2)
	if len(auth) != 2 || auth[0] != "Bearer" {
		return "", apperror.ErrAuthorizationToken
	}

	return auth[1], nil
}

func (au *Auth) Authenticate(req *http.Request) (*jwt.Token, error) {
	tokenString, err := au.parseHeader(req)
	if err != nil {
		return nil, apperror.ErrParseHeader
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(au.cfg.JwtSecret), nil
	})
	if err != nil {
		return nil, apperror.ErrParseToken
	}

	return token, nil
}

func (au *Auth) IsAdmin(token *jwt.Token) error {
	claims := token.Claims.(jwt.MapClaims)

	// Get user role
	role := claims["role"].(string)
	if role != "admin" {
		return apperror.ErrPermission
	}

	return nil
}
