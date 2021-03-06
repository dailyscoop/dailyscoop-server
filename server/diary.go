package server

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"

	"dailyscoop-backend/model"
)

func (s *Server) GetAllDiaries(c echo.Context) error {
	var diaries []model.Diary
	sortStr := c.QueryParam("sort")
	if sortStr != "" && (sortStr != "1" && sortStr != "-1") {
		return echo.NewHTTPError(http.StatusBadRequest, "정렬기준을 확인해주세요.")
	}
	var sort int
	var err error
	if sortStr == "" {
		sort = -1
	} else {
		sort, err = strconv.Atoi(sortStr)
		if err != nil {
			return nil
		}
	}
	content := c.QueryParam("search")
	if content == "" {
		diaries, err = s.ds.DiariesByUserID(c.Request().Context(), s.GetUserID(c), sort)
		if err != nil {
			return err
		}
	} else {
		diaries, err = s.ds.FindDiaries(c.Request().Context(), s.GetUserID(c), content, sort)
		if err != nil {
			return err
		}
	}
	type Diary struct {
		Content  string    `json:"content"`
		Image    string    `json:"image"`
		Date     time.Time `json:"date"`
		Emotions []string  `json:"emotions"`
		Theme    string    `json:"theme"`
	}
	resp := struct {
		Diaries []Diary `json:"diaries"`
	}{
		Diaries: []Diary{},
	}
	for _, diary := range diaries {
		resp.Diaries = append(resp.Diaries, Diary{
			Content:  diary.Content,
			Image:    diary.Image,
			Date:     diary.Date,
			Emotions: diary.Emotions,
			Theme:    diary.Theme,
		})
	}
	return c.JSON(http.StatusOK, resp)
}

func (s *Server) GetCalendar(c echo.Context) error {
	var req struct {
		Date string `query:"date"`
		Type string `query:"type"`
		Sort string `query:"sort"`
	}
	if err := c.Bind(&req); err != nil {
		return err
	}
	fmt.Println(req.Date + req.Type + req.Sort)
	if req.Sort != "" && (req.Sort != "1" && req.Sort != "-1") {
		return echo.NewHTTPError(http.StatusBadRequest, "정렬기준을 확인해주세요.")
	}
	var sort int
	var err error
	if req.Sort == "" {
		sort = -1
	} else {
		sort, err = strconv.Atoi(req.Sort)
		if err != nil {
			return nil
		}
	}
	if req.Date == "" || req.Type == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "파라미터가 올바르지 않습니다.")
	}
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return err
	}
	var diaries []model.Diary
	if req.Type == "monthly" || req.Type == "weekly" {
		diaries, err = s.ds.Calendar(c.Request().Context(), s.GetUserID(c), req.Type, date, sort)
		if err != nil {
			return err
		}
	} else {
		return echo.NewHTTPError(http.StatusBadRequest, "존재하지 않는 타입입니다.")
	}
	type Diary struct {
		Content  string    `json:"content"`
		Image    string    `json:"image"`
		Date     time.Time `json:"date"`
		Emotions []string  `json:"emotions"`
		Theme    string    `json:"theme"`
	}
	resp := struct {
		Diaries []Diary `json:"diaries"`
	}{
		Diaries: []Diary{},
	}
	for _, diary := range diaries {
		resp.Diaries = append(resp.Diaries, Diary{
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
	if dateString == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "파라미터가 올바르지 않습니다.")
	}
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
			return echo.NewHTTPError(http.StatusNotFound, "해당 날짜에 일기가 존재하지 않습니다.")
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
		Date     string
		Theme    string
	}
	if err := c.Bind(&req); err != nil {
		return err
	}
	if req.Content == "" || req.Image == "" || len(req.Emotions) == 0 || req.Date == "" || req.Theme == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "파라미터가 올바르지 않습니다.")
	}
	isThemeExists, err := s.ds.ThemeExists(c.Request().Context(), req.Theme)
	if err != nil {
		return err
	}
	if !isThemeExists {
		return echo.NewHTTPError(http.StatusBadRequest, "존재하지 않는 테마입니다.")
	}
	for _, emotion := range req.Emotions {
		isEmotionExists, err := s.ds.EmotionExists(c.Request().Context(), emotion)
		if err != nil {
			return err
		}
		if !isEmotionExists {
			return echo.NewHTTPError(http.StatusBadRequest, "존재하지 않는 감정입니다.")
		}
	}
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return err
	}
	diary := model.Diary{
		Content:  req.Content,
		Image:    req.Image,
		Emotions: req.Emotions,
		UserID:   s.GetUserID(c),
		Date:     date,
		Theme:    req.Theme,
	}
	if err := s.ds.WriteDiary(c.Request().Context(), diary); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, echo.Map{
		"message": "일기를 작성했습니다.",
	})
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
	return c.JSON(http.StatusOK, echo.Map{
		"message": "일기를 삭제했습니다.",
	})
}

func (s *Server) CountDiaries(c echo.Context) error {
	userID := s.GetUserID(c)
	typ := c.QueryParam("type")
	if typ == "" || (typ != "weekly" && typ != "monthly" && typ != "yearly") {
		return echo.NewHTTPError(http.StatusBadRequest, "타입이 잘못되었습니다.")
	}
	if c.QueryParam("date") == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "날짜를 입력해주세요.")
	}
	date, err := time.Parse("2006-01-02", c.QueryParam("date"))
	if err != nil {
		return err
	}
	diaryCount, dayCount, err := s.ds.CountDiaries(c.Request().Context(), typ, date, userID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, echo.Map{
		"diary_count": diaryCount,
		"day_count":   dayCount,
	})
}

func (s *Server) CountEmotions(c echo.Context) error {
	userID := s.GetUserID(c)
	typ := c.QueryParam("type")
	if typ == "" || (typ != "monthly" && typ != "yearly") {
		return echo.NewHTTPError(http.StatusBadRequest, "타입이 잘못되었습니다.")
	}
	if c.QueryParam("date") == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "날짜를 입력해주세요.")
	}
	date, err := time.Parse("2006-01-02", c.QueryParam("date"))
	if err != nil {
		return err
	}
	emotions, err := s.ds.CountEmotions(c.Request().Context(), userID, typ, date)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, echo.Map{
		"emotions": emotions,
	})
}
