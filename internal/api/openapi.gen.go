package api

import "github.com/labstack/echo/v5"

// ServerInterface mirrors the handlers generated from openapi.yaml.
// It is intentionally small and can be replaced by oapi-codegen output.
type ServerInterface interface {
	GetHealth(ctx *echo.Context) error
	ListPlayers(ctx *echo.Context) error
	CreatePlayer(ctx *echo.Context) error
	DeletePlayer(ctx *echo.Context, playerID string) error
	ListMatches(ctx *echo.Context) error
	CreateMatch(ctx *echo.Context) error
	DeleteMatch(ctx *echo.Context, matchID string) error
	GetLeaderboard(ctx *echo.Context) error
}

func RegisterHandlers(e *echo.Echo, si ServerInterface) {
	e.GET("/health", si.GetHealth)
	e.GET("/players", si.ListPlayers)
	e.POST("/players", si.CreatePlayer)
	e.DELETE("/players/:playerId", func(ctx *echo.Context) error {
		return si.DeletePlayer(ctx, ctx.Param("playerId"))
	})
	e.GET("/matches", si.ListMatches)
	e.POST("/matches", si.CreateMatch)
	e.DELETE("/matches/:matchId", func(ctx *echo.Context) error {
		return si.DeleteMatch(ctx, ctx.Param("matchId"))
	})
	e.GET("/leaderboard", si.GetLeaderboard)
}
