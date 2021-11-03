package server

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"dailyscoop-backend/config"
	"dailyscoop-backend/service"
)

type Server struct {
	*echo.Echo
	cfg config.Config
	us  *service.UserService
	ds  *service.DiaryService
}

func NewServer(cfg config.Config, us *service.UserService, ds *service.DiaryService) *Server {
	s := &Server{
		Echo: echo.New(),
		cfg:  cfg,
		us:   us,
		ds:   ds,
	}
	s.Use(middleware.Logger())
	s.Use(middleware.Recover())
	return s
}

func (s *Server) RegisterRoutes() {
	jwtConfig := middleware.JWTConfig{
		Claims:     &jwtCustomClaims{},
		SigningKey: []byte("secret"),
	}
	api := s.Group("/api")

	api.POST("/login", s.Login)
	api.POST("/signup", s.SignUp)

	diaries := api.Group("/diaries")
	diaries.Use(middleware.JWTWithConfig(jwtConfig))
	diaries.GET("", s.GetDiaries)
	diaries.POST("", s.CreateDiary)
	diaries.GET("/:date", s.GetDiary)
}
