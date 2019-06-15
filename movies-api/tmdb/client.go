package tmdb

import (
	"encoding/json"
	"fmt"

	"net/http"
	"strconv"
	"strings"

	"github.com/drgomesp/frames/movies-api/store"
	log "github.com/sirupsen/logrus"
)

const (
	TMBdURI = "https://api.themoviedb.org/3"
)

type UpcomingResponse struct {
	Data []MovieListItem `json:"data"`
}

type UpcomingPage struct {
	Page         int             `json:"page"`
	Results      []MovieListItem `json:"results"`
	TotalPages   int             `json:"total_pages"`
	TotalResults int             `json:"total_results"`
}

type MovieListItem struct {
	PosterPath       string  `json:"poster_path"`
	Adult            bool    `json:"adult"`
	Overview         string  `json:"overview"`
	ReleaseDate      string  `json:"release_date"`
	GenreIDs         []int   `json:"genre_ids"`
	ID               int     `json:"id"`
	OriginalTitle    string  `json:"original_title"`
	OriginalLanguage string  `json:"original_language"`
	Title            string  `json:"title"`
	BackdropPath     string  `json:"backdrop_path"`
	Popularity       float32 `json:"popularity"`
	VoteCount        int     `json:"vote_count"`
	Video            bool    `json:"video"`
	VoteAverage      float32 `json:"vote_average"`
}

// TODO: create a warm-up function to load all records from upcoming endpoint and put them into a cache store
// TODO: read from cache, otherwise call warm-up and then read from cache again

type Client struct {
	cache  store.Cache
	client *http.Client

	upcomingPages int
}

func NewClient(apiKey string) (*Client, error) {
	return &Client{
		cache:  store.NewRedisCache(),
		client: &http.Client{},
	}, nil
}

func (c *Client) WarmupUpcoming() error {
	log.Info("Warming-up upcoming movies...")

	var (
		err          error
		currentPage  *UpcomingPage
		pagesToFetch int
	)

	if currentPage, err = c.fetchUpcomingPage(1); err == nil {
		if r, err := json.Marshal(currentPage.Results); err == nil {
			key := fmt.Sprintf("upcoming/%d", 1)
			log.Debugf("cache.Set(%v)", key)
			c.cache.Set(key, string(r))
		}
	}

	c.upcomingPages = currentPage.TotalPages
	pagesToFetch = currentPage.TotalPages - 1

	for pagesToFetch > 1 {
		page, err := c.fetchUpcomingPage(pagesToFetch)

		if err != nil {
			return err
		}

		go c.storeUpcomingPage(pagesToFetch, page)
		pagesToFetch--
	}

	return nil
}

func (c *Client) GetUpcoming() (*UpcomingResponse, error) {
	var (
		response = &UpcomingResponse{Data: make([]MovieListItem, 0)}
	)

	for i := 1; i < c.upcomingPages; i++ {
		key := fmt.Sprintf("upcoming/%d", i)

		var items []MovieListItem
		err := json.Unmarshal([]byte((c.cache.Get(key))), &items)

		if err != nil {
			return nil, err
		}

		response.Data = append(response.Data, items...)
	}

	return response, nil
}

func (c *Client) fetchUpcomingPage(pageNumber int) (*UpcomingPage, error) {
	var (
		err  error
		url  strings.Builder
		req  *http.Request
		resp *http.Response
		page UpcomingPage
	)

	url.WriteString(TMBdURI)
	url.WriteString("/movie/upcoming?api_key=1f54bd990f1cdfb230adb312546d765d&language=en-US&page=")
	url.WriteString(strconv.Itoa(pageNumber))

	if req, err = http.NewRequest("GET", url.String(), nil); err != nil {
		return nil, err
	}

	if resp, err = c.client.Do(req); err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	// Use json.Decode for reading streams of JSON data
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		return nil, err
	}

	return &page, nil
}

func (c *Client) storeUpcomingPage(pageNumber int, page *UpcomingPage) error {
	r, err := json.Marshal(page.Results)

	if err != nil {
		return err
	}

	key := fmt.Sprintf("upcoming/%v", pageNumber)
	log.Debugf("cache.Set(%v)", key)

	c.cache.Set(key, string(r))

	return nil
}
