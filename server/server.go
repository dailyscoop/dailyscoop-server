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
	fs  *service.FavoriteService
	as  *service.AWSService
}

func NewServer(cfg config.Config, us *service.UserService, ds *service.DiaryService, fs *service.FavoriteService, as *service.AWSService) *Server {
	s := &Server{
		Echo: echo.New(),
		cfg:  cfg,
		us:   us,
		ds:   ds,
		fs:   fs,
		as:   as,
	}
	s.Use(middleware.Logger())
	s.Use(middleware.Recover())
	return s
}

func (s *Server) RegisterRoutes() {
	api := s.Group("/api")

	api.POST("/login", s.Login)
	api.POST("/signup", s.SignUp)
	api.POST("/image", s.ImageUpload)

	user := api.Group("/user")
	user.Use(middleware.JWTWithConfig(middleware.JWTConfig{
		Claims:     &jwtCustomClaims{},
		SigningKey: []byte(s.cfg.Server.Secret),
	}))

	user.GET("", s.GetUserInfo)
	user.DELETE("", s.DeleteUser)
	user.PUT("/change_password", s.ChangePassword)
	user.PUT("/change_nickname", s.ChangeNickname)
	user.PUT("/set_image", s.SetProfileImage)

	diaries := api.Group("/diaries")
	diaries.Use(middleware.JWTWithConfig(middleware.JWTConfig{
		Claims:     &jwtCustomClaims{},
		SigningKey: []byte(s.cfg.Server.Secret),
	}))

	diaries.GET("", s.GetAllDiaries)
	diaries.GET("/calendar", s.GetCalendar)
	diaries.POST("", s.CreateDiary)
	diaries.GET("/:date", s.GetDiary)
	diaries.DELETE("/:date", s.DeleteDiary)
	diaries.GET("/count", s.CountDiaries)
	diaries.GET("/emotions", s.CountEmotions)

	favorites := api.Group("/favorites")
	favorites.Use(middleware.JWTWithConfig(middleware.JWTConfig{
		Claims:     &jwtCustomClaims{},
		SigningKey: []byte(s.cfg.Server.Secret),
	}))

	favorites.GET("", s.GetFavorites)
	favorites.POST("", s.AddFavorite)
	favorites.DELETE("", s.DeleteFavorite)
}
