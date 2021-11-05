package server

import (
	"errors"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"

	"dailyscoop-backend/model"
)

func (s *Server) GetUserID(c echo.Context) string {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(*jwtCustomClaims)
	id := claims.ID
	return id
}

func (s *Server) GetDiaries(c echo.Context) error {
	diaries, err := s.ds.DiariesByUserID(c.Request().Context(), s.GetUserID(c))
	if err != nil {
		return err
	}
	type Diary struct {
		Content  string    `json:"content"`
		Image    string    `json:"image"`
		Date     time.Time `json:"date"`
		Emotions []string  `json:"emotions"`
		Theme    string    `json:"theme"`
	}
	resp := []Diary{}
	for _, diary := range diaries {
		resp = append(resp, Diary{
			Content:  diary.Content,
			Image:    diary.Image,
			Date:     diary.Date,
			Emotions: diary.Emotions,
			Theme:    diary.Theme,
		})
	}
	return c.JSON(http.StatusOK, resp)
}

func (s *Server) GetDiary(c echo.Context) error {
	dateString := c.Param("date")
	userID := s.GetUserID(c)
	date, err := time.Parse("2006-01-02", dateString)
	if err != nil {
		return err
	}
	diary, err := s.ds.DiaryByUserIDAndDate(c.Request().Context(), userID, date)
	type Diary struct {
		Content  string    `json:"content"`
		Image    string    `json:"image"`
		Date     time.Time `json:"date"`
		Emotions []string  `json:"emotions"`
		Theme    string    `json:"theme"`
	}
	resp := Diary{
		Content:  diary.Content,
		Image:    diary.Image,
		Date:     diary.Date,
		Emotions: diary.Emotions,
		Theme:    diary.Theme,
	}
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusNotFound, "diary not found for given date")
		}
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

func (s *Server) CreateDiary(c echo.Context) error {
	var req struct {
		Content  string
		Image    string
		Emotions []string
		Theme    string
	}
	if err := c.Bind(&req); err != nil {
		return err
	}
	diary := model.Diary{
		Content:  req.Content,
		Image:    req.Image,
		Emotions: req.Emotions,
		UserID:   s.GetUserID(c),
		Date:     time.Now(),
		Theme:    req.Theme,
	}
	err := s.ds.WriteDiary(c.Request().Context(), diary)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return echo.NewHTTPError(http.StatusBadRequest, "theme does not exist")
	} else if err != nil {
		return err
	}
	return c.NoContent(http.StatusOK)
}

func (s *Server) DeleteDiary(c echo.Context) error {
	dateString := c.Param("date")
	date, err := time.Parse("2006-01-02", dateString)
	if err != nil {
		return err
	}
	if err := s.ds.DeleteDiary(c.Request().Context(), s.GetUserID(c), date); err != nil {
		return err
	}
	return c.NoContent(http.StatusOK)
}
