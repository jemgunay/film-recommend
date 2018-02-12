package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

var (
	rootPath    = os.Getenv("GOPATH") + "/src/github.com/jemgunay/film-recommend"
	dbWrapper   DBWrapper
	recommender Recommender
)

func main() {
	// init recommender
	recommender = NewRecommender()
	// init DB
	dbWrapper = NewDatabase()
	// init server
	router := mux.NewRouter()

	// routes
	router.HandleFunc("/", searchHandler).Methods(http.MethodGet)
	router.HandleFunc("/users", userHandler).Methods(http.MethodGet)
	router.HandleFunc("/watched", watchedHandler).Methods(http.MethodGet)
	router.HandleFunc("/recommend", recommendHandler).Methods(http.MethodGet)

	// file server
	staticFileHandler := http.StripPrefix("/static/", http.FileServer(http.Dir(rootPath+"/static/")))
	router.Handle(`/static/{rest:[a-zA-Z0-9=\-\/._]+}`, staticFileHandler)

	port := 8005
	host := "127.0.0.1"
	server := &http.Server{
		Handler:      router,
		Addr:         host + ":" + fmt.Sprintf("%d", port),
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
	}

	// listen for HTTP requests
	log.Printf("starting HTTP server on port %d", port)
	if err := server.ListenAndServe(); err != nil {
		log.Println(err)
	}
}

// Get home HTML.
func searchHandler(w http.ResponseWriter, r *http.Request) {
	htmlResult := completeTemplate("/dynamic/search.html", nil)

	httpRespond(w, r, htmlResult, http.StatusOK)
}

// Get a recommendation for a specific user.
func recommendHandler(w http.ResponseWriter, r *http.Request) {
	// parse params
	params, err := getDataParams(r)
	if err != nil {
		httpRespond(w, r, "", http.StatusBadRequest)
		return
	}

	if params["user_id"] == "" {
		httpRespond(w, r, "no userid provided", http.StatusBadRequest)
		return
	}

	httpRespond(w, r, "recommendation here", http.StatusOK)
}

// Get home HTML.
func watchedHandler(w http.ResponseWriter, r *http.Request) {
	params, err := getDataParams(r)
	if err != nil {
		httpRespond(w, r, "", http.StatusBadRequest)
		return
	}

	if params["user_id"] == "" {
		httpRespond(w, r, "no userid provided", http.StatusBadRequest)
		return
	}
	if params["film_id"] == "" {
		httpRespond(w, r, "no userid provided", http.StatusBadRequest)
		return
	}

	httpRespond(w, r, "ok response", http.StatusOK)
}

// Get all user data.
func userHandler(w http.ResponseWriter, r *http.Request) {
	user, err := dbWrapper.perform("GetUserByName", "rob")
	fmt.Println(user, err)

	httpRespond(w, r, user, http.StatusOK)
}
