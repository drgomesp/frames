package handler

import (
	"net/http"

	"github.com/drgomesp/frames/movies-api/tmdb"
	"github.com/labstack/echo"
)

type MoviesHandler struct {
	client *tmdb.Client
}

func NewMoviesHandler(client *tmdb.Client) (*MoviesHandler, error) {
	return &MoviesHandler{
		client,
	}, nil
}

func (h *MoviesHandler) UpcomingHandler(c echo.Context) error {
	upcoming, err := h.client.GetUpcoming()

	if err != nil {
		panic(err)
	}

	return c.JSON(http.StatusOK, upcoming)
}
