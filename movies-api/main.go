package main

import (
	"os"

	"github.com/drgomesp/frames/movies-api/handler"
	"github.com/drgomesp/frames/movies-api/tmdb"
	"github.com/joho/godotenv"
	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		panic("could not load .env file")
	}

	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)

	logLevel, err := log.ParseLevel(os.Getenv("LOG_LEVEL"))

	if err != nil {
		panic(err)
	}

	log.SetLevel(logLevel)
}

func main() {
	e := echo.New()

	client, err := tmdb.NewClient(os.Getenv("TMDB_API_KEY"))
	if err != nil {
		log.Error(err)
	}

	client.WarmupGenres()
	go client.WarmupUpcoming()
	moviesHandler, err := handler.NewMoviesHandler(client)
	if err != nil {
		log.Error(err)
	}

	e.GET("/", moviesHandler.UpcomingHandler)

	log.Fatal(e.Start(":1323"))
}
