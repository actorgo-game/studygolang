package controller

import echo "github.com/labstack/echo/v4"

func RegisterRoutes(g *echo.Group) {
	new(InstallController).RegisterRoute(g)
	new(WebsocketController).RegisterRoute(g)
	new(OAuthController).RegisterRoute(g)
	new(CaptchaController).RegisterRoute(g)
	new(FeedController).RegisterRoute(g)
	new(ImageController).RegisterRoute(g)
}
