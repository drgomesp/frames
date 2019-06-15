package main

import (
	"os"

	"github.com/drgomesp/frames/movies-api/handler"
	"github.com/drgomesp/frames/movies-api/tmdb"
	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

func main() {
	e := echo.New()

	client, err := tmdb.NewClient(os.Getenv("TMDB_API_KEY"))
	if err != nil {
		log.Error(err)
	}

	go client.WarmupUpcoming()
	moviesHandler, err := handler.NewMoviesHandler(client)
	if err != nil {
		log.Error(err)
	}

	e.GET("/", moviesHandler.UpcomingHandler)

	log.Fatal(e.Start(":1323"))
}
