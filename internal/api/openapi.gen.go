package api

import "github.com/labstack/echo/v4"

// ServerInterface mirrors the handlers generated from openapi.yaml.
// It is intentionally small and can be replaced by oapi-codegen output.
type ServerInterface interface {
	GetHealth(ctx echo.Context) error
	ListPlayers(ctx echo.Context) error
	CreatePlayer(ctx echo.Context) error
	ListMatches(ctx echo.Context) error
	CreateMatch(ctx echo.Context) error
	GetLeaderboard(ctx echo.Context) error
}

func RegisterHandlers(e *echo.Echo, si ServerInterface) {
	e.GET("/health", si.GetHealth)
	e.GET("/players", si.ListPlayers)
	e.POST("/players", si.CreatePlayer)
	e.GET("/matches", si.ListMatches)
	e.POST("/matches", si.CreateMatch)
	e.GET("/leaderboard", si.GetLeaderboard)
}
