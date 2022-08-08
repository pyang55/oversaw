package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var ip string
var token string
var rating string
var minimal bool
var apikey string
var overseerUrl string

//xml structs
type Directories struct {
	XMLName     xml.Name    `xml:"MediaContainer"`
	Directories []Directory `xml:"Directory"`
}

type Directory struct {
	Type      string `xml:"type,attr"`
	Key       string `xml:"key,attr"`
	Title     string `xml:"title,attr"`
	Year      string `xml:"year,attr"`
	RatingKey string `xml:"ratingKey,attr"`
}

type Videos struct {
	XMLName xml.Name `xml:"MediaContainer"`
	Videos  []Video  `xml:"Video"`
}

type Video struct {
	Key       string `xml:"key,attr"`
	Rating    string `xml:"audienceRating,attr"`
	Type      string `xml:"type,attr"`
	Title     string `xml:"title,attr"`
	Year      string `xml:"year,attr"`
	RatingKey string `xml:"ratingKey,attr"`
}

//json structs for overseer
type Results struct {
	Results []Result `json:"results,omitempty"`
}

type Result struct {
	ReleaseDate  string `json:"releaseDate,omitempty"`
	Id           int    `json:"id,omitempty"`
	FirstAirDate string `json:"firstAirDate,omitempty"`
	MediaType    string `json:"mediaType,omitempty"`
}

func GetHttpRequests(url string) io.ReadCloser {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		// handle err
	}

	resp, err := client.Do(req)
	if err != nil {
		// handle err
	}
	//defer resp.Body.Close()

	return resp.Body
}

func PUTHttpRequests(url string) io.ReadCloser {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		// handle err
	}

	resp, err := client.Do(req)
	if err != nil {
		// handle err
	}
	//defer resp.Body.Close()

	return resp.Body
}

// Checks your watchlist for tv shows
func GetShowsWatchlist(overseerUrl string, token string) {
	url := fmt.Sprintf("https://metadata.provider.plex.tv/library/sections/watchlist/all?X-Plex-Token=%s", token)
	resp := GetHttpRequests(url)
	byteValue, _ := ioutil.ReadAll(resp)
	var vid Directories
	d := ""
	err := xml.Unmarshal(byteValue, &vid)
	if err != nil {
		fmt.Println(err)
	}

	for i := 0; i < len(vid.Directories); i++ {
		if vid.Directories[i].Type == "show" {
			d = "tv"
		} else {
			d = vid.Directories[i].Type
		}
		fmt.Println("Searching for", vid.Directories[i].Title)
		resp = OverseerSearch(overseerUrl, vid.Directories[i].Title, apikey)
		byteValue, _ = ioutil.ReadAll(resp)
		var r Results
		err = json.Unmarshal(byteValue, &r)
		if err != nil {
			fmt.Println(err)
		}
		if len(r.Results) == 1 {
			OverseerRequestShows(overseerUrl, d, r.Results[0].Id, apikey)
			rk := fmt.Sprintf("https://metadata.provider.plex.tv/actions/removeFromWatchlist?ratingKey=%s&X-Plex-Token=%s", vid.Directories[i].RatingKey, token)
			fmt.Println("Removing from watchlist")
			PUTHttpRequests(rk)
		}
		for z := 0; z < len(r.Results); z++ {
			if (strings.Contains(r.Results[z].ReleaseDate, vid.Directories[i].Year) || strings.Contains(r.Results[z].FirstAirDate, vid.Directories[i].Year)) && (r.Results[z].MediaType == d) {
				OverseerRequestShows(overseerUrl, d, r.Results[z].Id, apikey)
				rk := fmt.Sprintf("https://metadata.provider.plex.tv/actions/removeFromWatchlist?ratingKey=%s&X-Plex-Token=%s", vid.Directories[i].RatingKey, token)
				fmt.Println("Removing from watchlist")
				PUTHttpRequests(rk)
			}
		}
	}
}

// Checks your watchlist for movies
func GetMoviesWatchlist(overseerUrl string, token string, apikey string) {
	url := fmt.Sprintf("https://metadata.provider.plex.tv/library/sections/watchlist/all?X-Plex-Token=%s", token)
	resp := GetHttpRequests(url)
	byteValue, _ := ioutil.ReadAll(resp)
	var vid Videos
	err := xml.Unmarshal(byteValue, &vid)
	if err != nil {
		fmt.Println(err)
	}
	for i := 0; i < len(vid.Videos); i++ {
		fmt.Println("Searching for", vid.Videos[i].Title)
		resp = OverseerSearch(overseerUrl, vid.Videos[i].Title, apikey)
		byteValue, _ = ioutil.ReadAll(resp)
		var r Results
		err = json.Unmarshal(byteValue, &r)
		if err != nil {
			fmt.Println(err)
		}
		if len(r.Results) == 1 {
			OverseerRequestMovies(overseerUrl, "movie", r.Results[0].Id, apikey)
			rk := fmt.Sprintf("https://metadata.provider.plex.tv/actions/removeFromWatchlist?ratingKey=%s&X-Plex-Token=%s", vid.Videos[i].RatingKey, token)
			fmt.Println("Removing from watchlist")
			PUTHttpRequests(rk)
		} else {
			for z := 0; z < len(r.Results); z++ {
				if strings.Contains(r.Results[z].ReleaseDate, vid.Videos[i].Year) && r.Results[z].MediaType == "movie" {
					OverseerRequestMovies(overseerUrl, "movie", r.Results[z].Id, apikey)
					rk := fmt.Sprintf("https://metadata.provider.plex.tv/actions/removeFromWatchlist?ratingKey=%s&X-Plex-Token=%s", vid.Videos[i].RatingKey, token)
					fmt.Println("Removing from watchlist")
					PUTHttpRequests(rk)
				}
			}
		}
	}
}

func OverseerSearch(overseerUrl string, title string, apikey string) io.ReadCloser {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	t := url.PathEscape(title)
	t = strings.Replace(t, ":", "%3A", -1)
	url := fmt.Sprintf("%s/api/v1/search?query=%s&page=1&language=en", overseerUrl, t)
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Add("x-api-key", apikey)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	//defer resp.Body.Close()
	return resp.Body
}

func OverseerRequestShows(overseerUrl string, media string, id int, apikey string) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	// we have to do this because overseer doesn't specify an "all" option
	seasons := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27}
	payload := map[string]interface{}{"mediaType": media, "mediaId": id, "seasons": seasons}
	byts, _ := json.Marshal(payload)
	fmt.Println(string(byts)) // {"id":1,"name":"zahid"}

	url := fmt.Sprintf("%s/api/v1/request", overseerUrl)
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(byts))
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("x-api-key", apikey)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	// defer resp.Body.Close()
	fmt.Println(resp.Status)
}

// Request movies on overseer
func OverseerRequestMovies(overseerUrl string, media string, id int, apikey string) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	payload := map[string]interface{}{"mediaType": media, "mediaId": id}
	byts, _ := json.Marshal(payload)

	url := fmt.Sprintf("%s/api/v1/request", overseerUrl)
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(byts))
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("x-api-key", apikey)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	// defer resp.Body.Close()
	fmt.Println(resp.Status)
}

//Start Watcher function, runs every 30 seconds
func StartWatcher(url string, token string, apikey string) {
	ticker := time.NewTicker(30 * time.Second)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				fmt.Println("Running a Overseer Search")
				GetShowsWatchlist(url, token)
				GetMoviesWatchlist(url, token, apikey)
			}
		}
	}()
}

// CLI COMMANDS
var grabmovies = &cobra.Command{
	Use:   "grab-movies",
	Short: "filter shows by rating or version",
	Run: func(cmd *cobra.Command, args []string) {
		GetMoviesWatchlist(overseerUrl, token, apikey)
		os.Exit(0)
	},
}

var grabshows = &cobra.Command{
	Use:   "grab-shows",
	Short: "filter shows by rating or version",
	Run: func(cmd *cobra.Command, args []string) {
		GetShowsWatchlist(overseerUrl, token)
		os.Exit(0)
	},
}

var start = &cobra.Command{
	Use:   "start-oversaw",
	Short: "filter shows by rating or version",
	Run: func(cmd *cobra.Command, args []string) {
		timeout := time.After(8 * time.Hour)
		ticker := time.NewTicker(30 * time.Second)
		for {
			select {
			case <-timeout:
				fmt.Println("Exiting")
				return
			case <-ticker.C:
				fmt.Println("Running a Overseer Search")
				GetShowsWatchlist(overseerUrl, token)
				GetMoviesWatchlist(overseerUrl, token, apikey)
			}
		}

	},
}

func main() {
	var rootCmd = &cobra.Command{Use: "oversaw"}
	grabmovies.Flags().StringVarP(&overseerUrl, "url", "u", os.Getenv("URL"), "Overseer url or ip:port")
	grabmovies.Flags().StringVarP(&token, "token", "t", os.Getenv("TOKEN"), "Plex Token, more information https://support.plex.tv/articles/204059436-finding-an-authentication-token-x-plex-token/")
	grabmovies.Flags().StringVarP(&apikey, "apikey", "a", os.Getenv("APIKEY"), "Overseer Apikey Token")

	grabshows.Flags().StringVarP(&overseerUrl, "url", "u", os.Getenv("URL"), "Overseer url or ip:port")
	grabshows.Flags().StringVarP(&token, "token", "t", os.Getenv("TOKEN"), "Plex Token, more information https://support.plex.tv/articles/204059436-finding-an-authentication-token-x-plex-token/")
	grabshows.Flags().StringVarP(&apikey, "apikey", "a", "", "Overseer Apikey Token")

	start.Flags().StringVarP(&overseerUrl, "url", "u", os.Getenv("URL"), "Overseer url or ip:port")
	start.Flags().StringVarP(&token, "token", "t", os.Getenv("TOKEN"), "Plex Token, more information https://support.plex.tv/articles/204059436-finding-an-authentication-token-x-plex-token/")
	start.Flags().StringVarP(&apikey, "apikey", "a", os.Getenv("APIKEY"), "Overseer Apikey Token")

	rootCmd.AddCommand(grabmovies)
	rootCmd.AddCommand(grabshows)
	rootCmd.AddCommand(start)

	rootCmd.Execute()
}
