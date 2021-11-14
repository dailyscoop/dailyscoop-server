package server

import (
	"errors"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"

	"github.com/labstack/echo/v4"
)

type jwtCustomClaims struct {
	ID string `json:"id"`
	jwt.StandardClaims
}

func (s *Server) Login(c echo.Context) error {
	var req struct {
		ID       string
		Password string
	}
	if err := c.Bind(&req); err != nil {
		return err
	}
	user, err := s.us.UserByID(c.Request().Context(), req.ID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusUnauthorized, "아이디나 비밀번호를 확인해주세요.")
		}
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "아이디나 비밀번호를 확인해주세요.")
	}

	claims := &jwtCustomClaims{
		user.ID,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 720).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	t, err := token.SignedString([]byte(s.cfg.Server.Secret))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{
		"token":    t,
		"nickname": user.Nickname,
	})
}

func (s *Server) SignUp(c echo.Context) error {
	var req struct {
		ID       string
		Password string
		Nickname string
	}
	if err := c.Bind(&req); err != nil {
		return err
	}
	if req.ID == "" || req.Password == "" || req.Nickname == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "파라미터가 올바르지 않습니다.")
	}
	ctx := c.Request().Context()
	_, err := s.us.UserByID(ctx, req.ID)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return err
	} else if err == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "이미 존재하는 아이디입니다.")
	}

	_, err = s.us.UserByNickname(ctx, req.Nickname)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return err
	} else if err == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "이미 존재하는 닉네임입니다.")
	}

	if err := s.us.RegisterUser(ctx, struct {
		ID           string
		Password     string
		Nickname     string
		ProfileImage string
	}{ID: req.ID, Password: req.Password, Nickname: req.Nickname, ProfileImage: ""}); err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, echo.Map{
		"message": "회원가입이 완료되었습니다.",
	})
}

func (s *Server) DeleteUser(c echo.Context) error {
	userID := s.GetUserID(c)
	if userID != c.Param("userID") {
		return echo.NewHTTPError(http.StatusBadRequest, "자신의 계정만 탈퇴할 수 있습니다.")
	}
	if err := s.us.DeleteUser(c.Request().Context(), userID); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, echo.Map{
		"message": "회원탈퇴가 완료되었습니다.",
	})
}

func (s *Server) ChangeNickname(c echo.Context) error {
	var req struct {
		NewNickname string
	}
	if err := c.Bind(&req); err != nil {
		return err
	}
	if req.NewNickname == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "변경할 닉네임을 입력해주세요.")
	}
	if err := s.us.UpdateNickname(c.Request().Context(), s.GetUserID(c), req.NewNickname); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, echo.Map{
		"message": "닉네임을 변경했습니다.",
	})
}

func (s *Server) ChangePassword(c echo.Context) error {
	var req struct {
		Password    string
		NewPassword string
	}
	if err := c.Bind(&req); err != nil {
		return err
	}
	if req.Password == "" || req.NewPassword == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "비밀번호와 새 비밀번호를 모두 입력해주세요.")
	}
	userID := s.GetUserID(c)
	user, err := s.us.UserByID(c.Request().Context(), userID)
	if err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "비밀번호가 일치하지 않습니다.")
	}
	if err := s.us.UpdatePassword(c.Request().Context(), userID, req.NewPassword); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, echo.Map{
		"message": "비밀번호가 변경되었습니다.",
	})
}

func (s *Server) GetUserInfo(c echo.Context) error {
	user, err := s.us.UserByID(c.Request().Context(), s.GetUserID(c))
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return echo.NewHTTPError(http.StatusBadRequest, "존재하지 않는 유저입니다.")
		}
		return err
	}
	userID := c.Param("userID")
	if user.ID != userID {
		return echo.NewHTTPError(http.StatusBadRequest, "자신의 정보만 열람할 수 있습니다.")
	}
	type resp struct {
		ID       string `json:"id"`
		Nickname string `json:"nickname"`
	}
	return c.JSON(http.StatusOK, resp{
		ID:       user.ID,
		Nickname: user.Nickname,
	})
}
