package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	baseURL = "https://api.themoviedb.org/3"
	apiKeys = []string{"9d35385bd6e30e1da8b4350e5be48b44"}

	currentID         = 0
	mostRecentID      = 0
	pendingIDRequests = make(map[int]bool)
	idFeed            = IDFeed{}
	counter           = StatCounter{store: make(map[string]int)}

	// 40 requests/10s = 240 requests/min
	maxRequests            = 40
	requestCountTimeout    = time.Millisecond * 10000 // add an extra 10 ms to reduce limit window collisions
	sampleFilmDataByteSize = 1314                     // size of interstellar response ( curl -I http://api.themoviedb.org/3/movie/157336?api_key=9d35385bd6e30e1da8b4350e5be48b44 )

	// define HTTP client with timeout
	httpClient = &http.Client{Timeout: time.Second * 10}
)

type MovieResult struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

func main() {
	// get latest film ID
	latestFilmURL := fmt.Sprintf("%v/movie/latest?api_key=%v", baseURL, apiKeys[0])
	response := request(ProcRequest{url: latestFilmURL, readBody: true})
	if response.err != nil {
		log.Printf("error fetching most recent film ID, error:[%v]", response.err)
		return
	}

	movieResult := MovieResult{}
	if err := json.Unmarshal([]byte(response.body), &movieResult); err != nil {
		log.Printf("error parsing most recent film ID, error:[%v]", err)
		return
	}

	// validate flags
	mostRecentID = movieResult.ID
	if currentID >= mostRecentID {
		log.Printf("starting ID must be smaller than the most recent ID")
		return
	}
	mostRecentID = 100

	// estimate download duration
	fmt.Printf("Most recent movie ID: %v\n", mostRecentID)
	timePerAPIKey := (mostRecentID / (maxRequests * 6)) / 60
	fmt.Printf("> Estimates assume all IDs correspond with valid films & memory estimates use an average response size:\n")
	if len(apiKeys) > 1 {
		fmt.Printf("Rough time estimate for only one API key: ~%v hours\n", timePerAPIKey)
	}
	fmt.Printf("Rough time estimate with the provided %v API keys: ~%v hours\n", len(apiKeys), timePerAPIKey/len(apiKeys))
	fmt.Printf("Rough raw memory consumption estimate: ~%.3f Mb\n", (float64(mostRecentID*sampleFilmDataByteSize))*math.Pow10(-6))
	fmt.Printf("\n--- scraper starting in 10 seconds ---\n\n")

	// start all requesters
	for _, key := range apiKeys {
		go NewTMDBRequester(key)
	}

	// prevent exit
	exitCh := make(chan struct{})
	exitTicker := time.NewTicker(time.Second)
	go func() {
		for {
			<-exitTicker.C
			if currentID >= mostRecentID && len(pendingIDRequests) == 0 {
				exitCh <- struct{}{}
			}
		}
	}()
	<-exitCh

	// output stats
	fmt.Printf("\n> Stats:\n%+v", counter.store)
}

// Count stats.
type StatCounter struct {
	store map[string]int
	sync.Mutex
}

func (s *StatCounter) increment(stat string) {
	s.Lock()
	s.store[stat]++
	s.Unlock()
}

// Safely fetch next ID & keep track of IDs currently being processed in order to restart if paused.
type IDFeed struct {
	idFetchLock   sync.Mutex
	idPendingLock sync.Mutex
}

func (f *IDFeed) getNextID() (filmID int) {
	f.idPendingLock.Lock()

	// check for failed IDs stuck in pending
	for key := range pendingIDRequests {
		if pendingIDRequests[key] == false {
			filmID = key
			break
		}
	}

	if filmID == 0 {
		f.idFetchLock.Lock()
		currentID++
		filmID = currentID
		f.idFetchLock.Unlock()
	}

	pendingIDRequests[filmID] = true
	f.idPendingLock.Unlock()

	return
}

func (f *IDFeed) removeIDFromPending(filmID int) {
	f.idPendingLock.Lock()
	delete(pendingIDRequests, filmID)
	f.idPendingLock.Unlock()
}

func (f *IDFeed) remarkIDForPending(filmID int) {
	f.idPendingLock.Lock()
	pendingIDRequests[filmID] = false
	f.idPendingLock.Unlock()
}

// Perform data requests whilst enforcing the maximum request cap enforced by TMDB.
type TMDBRequester struct {
	requestCount int
	ticker       *time.Ticker
}

func NewTMDBRequester(apiKey string) (requester TMDBRequester) {
	requester.ticker = time.NewTicker(requestCountTimeout)

	for {
		<-requester.ticker.C

		for i := 0; i < maxRequests; i++ {
			go func() {
				nextID := idFeed.getNextID()
				latestFilmURL := fmt.Sprintf("%v/movie/%v?api_key=%v", baseURL, nextID, apiKey)
				response := request(ProcRequest{url: latestFilmURL, readBody: true})
				if response.err != nil {
					log.Printf("error fetching most recent film ID, error:[%v]", response.err)
					return
				}

				switch response.status {
				// request was for a valid ID
				case http.StatusOK:
					movieResult := MovieResult{}
					if err := json.Unmarshal([]byte(response.body), &movieResult); err != nil {
						log.Printf("error parsing most recent film ID, error:[%v]", err)
						return
					}

					fmt.Printf("%v\n", movieResult.Title)
					idFeed.removeIDFromPending(movieResult.ID)

					// store data
					counter.increment("valid_ids")

				// request was for an invalid ID
				case http.StatusNotFound:
					fmt.Printf("%v ID is invalid\n", nextID)
					idFeed.removeIDFromPending(nextID)

					counter.increment("invalid_ids")

				// request made but TMDB limit reached - user used API key in separate browser call by accident?
				case http.StatusTooManyRequests:
					// retry? do this from the feed, return map values (which equal false, not true) before new ones?
					idFeed.remarkIDForPending(nextID)

					counter.increment("tmdb_limit_reached")
				}
			}()
		}
	}

	return
}

// Contains data required to make a request.
type ProcRequest struct {
	url         string
	method      string
	body        string
	readBody    bool
	contentType string
}

// Contains data representing a response.
type ProcResponse struct {
	status int
	body   string
	err    error
}

// Perform a request.
func request(req ProcRequest) (resp ProcResponse) {
	// prepare request
	request, err := http.NewRequest(req.method, req.url, strings.NewReader(req.body))
	if err != nil {
		return
	}

	if req.contentType != "" {
		request.Header.Set("Content-Type", req.contentType)
	}

	// perform request
	response, err := httpClient.Do(request)
	if err != nil {
		// request failed
		return
	}
	defer response.Body.Close()

	resp.status = response.StatusCode

	// read body if readBody param is true
	if req.readBody {
		responseBodyResult, err := ioutil.ReadAll(response.Body)
		if err != nil {
			// reading response body failed
			return
		}

		resp.body = string(responseBodyResult)
	}

	return
}
