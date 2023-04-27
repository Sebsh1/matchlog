//go:generate mockgen --source=service.go -destination=service_mock.go -package=authentication
package authentication

import (
	"context"
	"foosball/internal/user"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type Config struct {
	Secret   string
	Issuer   string
	Audience string
}

type Service interface {
	Login(ctx context.Context, email, password string) (valid bool, token string, err error)
	VerifyJWT(ctx context.Context, token string) (valid bool, claims *Claims, err error)
	Signup(ctx context.Context, email, username, password string) error
}

type ServiceImpl struct {
	config      Config
	userService user.Service
}

func NewService(config Config, userService user.Service) Service {
	return &ServiceImpl{
		config:      config,
		userService: userService,
	}
}

func (s *ServiceImpl) Login(ctx context.Context, email string, password string) (bool, string, error) {
	exists, user, err := s.userService.GetUserByEmail(ctx, email)
	if err != nil {
		return false, "", err
	}

	if !exists {
		return false, "", nil
	}

	hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return false, "", err
	}

	if err = bcrypt.CompareHashAndPassword(hashedPasswordBytes, []byte(user.Hash)); err != nil {
		return false, "", nil
	}

	token, err := s.generateJWT(user.Name, user.ID, user.OrganizationID, user.Admin)
	if err != nil {
		return false, "", err
	}

	return true, token, nil
}

func (s *ServiceImpl) Signup(ctx context.Context, email string, username string, password string) error {
	exists, _, err := s.userService.GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}

	if exists {
		return errors.New("email alreadyin use")
	}

	hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return err
	}

	if err = s.userService.CreateUser(ctx, email, username, string(hashedPasswordBytes)); err != nil {
		return err
	}

	return nil
}

func (s *ServiceImpl) VerifyJWT(ctx context.Context, token string) (bool, *Claims, error) {
	parsedToken, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.config.Secret), nil
	})
	if err != nil {
		return false, nil, err
	}

	if _, ok := parsedToken.Method.(*jwt.SigningMethodHMAC); !ok {
		return false, nil, errors.New("unexpected signing method")
	}

	if !parsedToken.Valid {
		return false, nil, errors.New("jwt is invalid")
	}

	claims, ok := parsedToken.Claims.(*Claims)
	if !ok {
		return false, nil, errors.New("failed to parse claims")
	}

	return true, claims, nil
}

func (s *ServiceImpl) generateJWT(name string, userID, organizationID uint, admin bool) (string, error) {
	tokenUUID, err := uuid.NewRandom()
	if err != nil {
		return "", errors.New("failed to generate uuid")
	}

	now := time.Now()

	standardClaims := jwt.StandardClaims{
		Id:        tokenUUID.String(),
		IssuedAt:  now.Unix(),
		NotBefore: now.Unix(),
		ExpiresAt: now.Add(6 * time.Hour).Unix(),
		Issuer:    s.config.Issuer,
		Audience:  s.config.Audience,
		Subject:   strconv.FormatUint(uint64(userID), 10),
	}

	claims := Claims{
		StandardClaims: standardClaims,
		Name:           name,
		UserID:         userID,
		OrganizationID: organizationID,
		Admin:          admin,
	}

	tokenUnsigned := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenSigned, err := tokenUnsigned.SignedString([]byte(s.config.Secret))
	if err != nil {
		return "", errors.Wrap(err, "failed to sign access token")
	}

	return tokenSigned, nil
}