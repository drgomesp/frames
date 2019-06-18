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
	PosterPath   string  `json:"poster_path"`
	ReleaseDate  string  `json:"release_date"`
	GenreIDs     []int   `json:"genre_ids"`
	Title        string  `json:"title"`
	BackdropPath string  `json:"backdrop_path"`
	Genres       []Genre `json:"genres"`
}

type Genre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Client struct {
	apiKey string
	cache  store.Cache
	client *http.Client

	upcomingPages int
}

func NewClient(apiKey string) (*Client, error) {
	return &Client{
		apiKey: apiKey,
		cache:  store.NewRedisCache(),
		client: &http.Client{},
	}, nil
}

type GenresResponse struct {
	Genres []Genre `json:"genres"`
}

func (c *Client) WarmupGenres() error {
	log.Info("Warming-up genres...")

	var (
		err            error
		url            strings.Builder
		req            *http.Request
		resp           *http.Response
		genresResponse GenresResponse
	)

	url.WriteString(fmt.Sprintf("%s/genre/movie/list?api_key=%s&language=en-US", TMBdURI, c.apiKey))

	if req, err = http.NewRequest("GET", url.String(), nil); err != nil {
		return err
	}

	if resp, err = c.client.Do(req); err != nil {
		return err
	}

	defer resp.Body.Close()

	// Use json.Decode for reading streams of JSON data
	if err := json.NewDecoder(resp.Body).Decode(&genresResponse); err != nil {
		return err
	}

	for _, g := range genresResponse.Genres {
		key := fmt.Sprintf("genres/%d", g.ID)
		log.Debugf("cache.Set(%v)", key)
		c.cache.Set(key, g.Name)
	}

	return nil
}

func (c *Client) WarmupUpcoming() error {
	log.Info("Warming-up upcoming movies...")

	var (
		err         error
		currentPage *UpcomingPage
	)

	if currentPage, err = c.fetchUpcomingPage(1); err == nil {
		if r, err := json.Marshal(currentPage.Results); err == nil {
			key := fmt.Sprintf("upcoming/%d", 1)
			log.Debugf("cache.Set(%v)", key)
			c.cache.Set(key, string(r))
		}
	}

	c.upcomingPages = currentPage.TotalPages

	count := 1
	for count <= c.upcomingPages {
		page, err := c.fetchUpcomingPage(count)

		if err != nil {
			return err
		}

		go c.storeUpcomingPage(count, page)
		count++
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

	url.WriteString(fmt.Sprintf("%s/movie/upcoming?api_key=%s&language=en-US&page=%s", TMBdURI, c.apiKey, strconv.Itoa(pageNumber)))

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
	var results []MovieListItem

	for _, r := range page.Results {
		for _, genreID := range r.GenreIDs {
			name := c.cache.Get(fmt.Sprintf("genres/%d", genreID))
			log.Debugf("cache.Get(%v)", genreID)

			r.Genres = append(r.Genres, Genre{
				genreID,
				name,
			})
		}

		results = append(results, r)
	}

	r, err := json.Marshal(results)
	log.Debug(string(r))

	if err != nil {
		return err
	}

	key := fmt.Sprintf("upcoming/%v", pageNumber)
	log.Debugf("cache.Set(%v)", key)

	c.cache.Set(key, string(r))

	return nil
}
