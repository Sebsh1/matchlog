package rest

import (
	"context"
	"fmt"
	"matchlog/internal/authentication"
	"matchlog/internal/invite"
	"matchlog/internal/leaderboard"
	"matchlog/internal/match"
	"matchlog/internal/organization"
	"matchlog/internal/rating"
	"matchlog/internal/rest/controllers"
	"matchlog/internal/rest/helpers"
	"matchlog/internal/statistic"
	"matchlog/internal/user"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Port int `mapstructure:"port" default:"8001"`
}

type Server struct {
	echo *echo.Echo
	port int
}

func NewServer(
	port int,
	logger *logrus.Logger,
	authService authentication.Service,
	userService user.Service,
	organizationService organization.Service,
	inviteService invite.Service,
	matchService match.Service,
	ratingService rating.Service,
	statisticService statistic.Service,
	leaderboardService leaderboard.Service,
) (*Server, error) {
	e := echo.New()

	e.HideBanner = true
	e.HidePort = true
	e.Validator = helpers.NewValidator()

	e.Use(
		middleware.Recover(),
		middleware.Logger(),
		middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: []string{"*"},
		}),
		middleware.GzipWithConfig(middleware.GzipConfig{
			Skipper: middleware.DefaultGzipConfig.Skipper,
		}),
	)

	root := e.Group("/api")

	controllers.Register(
		root,
		logger.WithField("module", "rest"),
		authService,
		userService,
		organizationService,
		inviteService,
		matchService,
		ratingService,
		statisticService,
		leaderboardService,
	)

	return &Server{
		echo: e,
		port: port,
	}, nil
}

func (s *Server) Start() error {
	return errors.Wrap(s.echo.Start(fmt.Sprintf("0.0.0.0:%d", s.port)), "Failed to start server")
}

func (s *Server) Shutdown(ctx context.Context) error {
	return errors.Wrap(s.echo.Shutdown(ctx), "Failed to shutdown server")
}
