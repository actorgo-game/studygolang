package middleware

import (
	"net/http"

	echo "github.com/labstack/echo/v4"
)

func HTTPError() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			if err := next(ctx); err != nil {
				if !ctx.Response().Committed {
					if he, ok := err.(*echo.HTTPError); ok {
						switch he.Code {
						case http.StatusNotFound:
							return ctx.JSON(http.StatusNotFound, map[string]interface{}{"ok": 0, "error": "页面不存在"})
						case http.StatusForbidden:
							return ctx.JSON(http.StatusForbidden, map[string]interface{}{"ok": 0, "error": "没有权限访问"})
						case http.StatusInternalServerError:
							return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": 0, "error": "服务器内部错误"})
						default:
							return err
						}
					}
				}
			}
			return nil
		}
	}
}
