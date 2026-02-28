package http

import (
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/sessions"
	echo "github.com/labstack/echo/v4"
	"github.com/polaris1119/config"
	"github.com/polaris1119/goutils"

	"github.com/studygolang/studygolang/internal/logic"
)

var Store = sessions.NewCookieStore([]byte(config.ConfigFile.MustValue("global", "cookie_secret")))

func SetLoginCookie(ctx echo.Context, username string) {
	Store.Options.HttpOnly = true

	session := GetCookieSession(ctx)
	if ctx.FormValue("remember_me") != "1" {
		session.Options = &sessions.Options{
			Path:     "/",
			HttpOnly: true,
		}
	}
	session.Values["username"] = username
	req := Request(ctx)
	resp := ResponseWriter(ctx)
	session.Save(req, resp)
}

func GetCookieSession(ctx echo.Context) *sessions.Session {
	session, _ := Store.Get(Request(ctx), "user")
	return session
}

func Request(ctx echo.Context) *http.Request {
	return ctx.Request()
}

func ResponseWriter(ctx echo.Context) http.ResponseWriter {
	return ctx.Response()
}

func CheckIsHttps(ctx echo.Context) bool {
	isHttps := goutils.MustBool(ctx.Request().Header.Get("X-Https"))
	if logic.WebsiteSetting.OnlyHttps {
		isHttps = true
	}

	return isHttps
}

const (
	TokenSalt       = "b3%JFOykZx_golang_polaris"
	NeedReLoginCode = 600
)

func ParseToken(token string) (int, bool) {
	if len(token) < 32 {
		return 0, false
	}

	pos := strings.LastIndex(token, "uid")
	if pos == -1 {
		return 0, false
	}
	return goutils.MustInt(token[pos+3:]), true
}

func ValidateToken(token string) bool {
	_, ok := ParseToken(token)
	if !ok {
		return false
	}

	expireTime := time.Unix(goutils.MustInt64(token[:10]), 0)
	if time.Now().Before(expireTime) {
		return true
	}
	return false
}

func GenToken(uid int) string {
	expireTime := time.Now().Add(30 * 24 * time.Hour).Unix()

	buffer := goutils.NewBuffer().Append(expireTime).Append(uid).Append(TokenSalt)

	md5 := goutils.Md5(buffer.String())

	buffer = goutils.NewBuffer().Append(expireTime).Append(md5).Append("uid").Append(uid)
	return buffer.String()
}

func AccessControl(ctx echo.Context) {
	ctx.Response().Header().Add("Access-Control-Allow-Origin", "*")
}
