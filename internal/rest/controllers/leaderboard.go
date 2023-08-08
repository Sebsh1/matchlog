package controllers

import (
	"matchlog/internal/leaderboard"
	"matchlog/internal/rest/handlers"
	"matchlog/internal/rest/helpers"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h *Handlers) GetLeaderboard(c handlers.AuthenticatedContext) error {
	type getLeaderboardRequest struct {
		TopX            int                         `query:"topX" validate:"required,gt=0,lte=50"`
		LeaderboardType leaderboard.LeaderboardType `query:"type" validate:"required,oneof=wins streak rating"`
	}

	type getLeaderboardResponse struct {
		Leaderboard leaderboard.Leaderboard `json:"leaderboard"`
	}

	ctx := c.Request().Context()

	req, err := helpers.Bind[getLeaderboardRequest](c)
	if err != nil {
		return echo.ErrBadRequest
	}

	leaderboard, err := h.leaderboardService.GetLeaderboard(ctx, c.Claims.OrganizationID, req.TopX, req.LeaderboardType)
	if err != nil {
		h.logger.WithError(err).Error("failed to get leaderboard")
		return echo.ErrInternalServerError
	}

	response := getLeaderboardResponse{
		Leaderboard: *leaderboard,
	}

	return c.JSON(http.StatusOK, response)
}
