package main

import (
	"encoding/json"
	"log"
	"fmt"
	"net/http"
	
	"github.com/labstack/echo"
)

func main() {
	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		fetchUpcomingMovies()
		return c.String(http.StatusOK, "Hello, World!")
	})

	e.Logger.Fatal(e.Start(":1323"))
}

type UpcomingMoviesResponse struct {
	Results []MovieListResult `json:"results"`
	TotalPages int `json:"total_pages"`
	TotalResults int `json:"total_results"`
}

type MovieListResult struct {
	PosterPath string `json:"poster_path"`
	Adult bool `json:"adult"`
	Overview string `json:"overview"`
	ReleaseDate string `json:"release_date"`
	GenreIDs []int `json:"genre_ids"`
	ID int `json:"id"`
	OriginalTitle string `json:"original_title"`
	OriginalLanguage string `json:"original_language"`
	Title string `json:"title"`
	BackdropPath string `json:"backdrop_path"`
	Popularity float32 `json:"popularity"`
	VoteCount int `json:"vote_count"`
	Video bool `json:"video"`
	VoteAverage float32 `json:"vote_average"`
}

func fetchUpcomingMovies() {
	url := fmt.Sprintf("https://api.themoviedb.org/3/movie/upcoming?api_key=1f54bd990f1cdfb230adb312546d765d&language=en-US&page=1")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("NewRequest: ", err)
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Do: ", err)
		return
	}

	defer resp.Body.Close()

	var r UpcomingMoviesResponse

	// Use json.Decode for reading streams of JSON data
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		log.Println(err)
	}

	log.Println(r.TotalPages)
	log.Println(r.TotalResults)
}
