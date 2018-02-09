package film_recommendation_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/muesli/regommend"
)

var films = regommend.Table("films")

func main() {
	parseDB()

	// set up HTTP endpoints
	http.HandleFunc("/recommend", getRecommendation)
	if err := http.ListenAndServe("127.0.0.1:8000", nil); err != nil {
		fmt.Printf("Error starting HTTP endpoints, error:[%v]", err)
	}

}

// Produce a recommendation for a user.
func getRecommendation(w http.ResponseWriter, r *http.Request) {
	// parse params
	params, err := getDataParams(r)
	if err != nil {
		httpRespond(w, r, "", http.StatusBadRequest)
		return
	}

	if params["userid"] == "" {
		httpRespond(w, r, "no userid provided", http.StatusBadRequest)
		return
	}

	friendsOnly := params["friends"] == "true"
	friendsOnly = friendsOnly

	// get recommendations
	recs, _ := films.Recommend(12345)
	for result := 0; result < 10; result++ {
		fmt.Printf("[%v]: Recommending %v with score %v\n", result, recs[result].Key, recs[result].Distance)
	}

	httpRespond(w, r, "recommendation here", http.StatusOK)
}

// Pull film data from DB and parse into regommender format.
func parseDB() {
	films.Flush()

	watchedFilms := make(map[interface{}]float64)
	watchedFilms["film_id"] = 4.0

	films.Add(12345, watchedFilms)
}

// Respond to HTTP request clean up request body.
func httpRespond(w http.ResponseWriter, r *http.Request, response interface{}, status int) {
	w.WriteHeader(status)

	// write response
	if _, err := fmt.Fprintf(w, "%v", response); err != nil {
		fmt.Printf("error writing response, error:[%v] response:[%v (%v)]", err, response, status)
	}

	r.Body.Close()
}

// Get & validate a parameter from the request body/data.
func getDataParams(r *http.Request) (params map[string]string, err error) {
	params = make(map[string]string)

	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return params, fmt.Errorf("error reading request body, query=[%v]", r.URL.RawQuery)
	}

	// collect all parameter key:value pairs
	paramsPairs, err := url.ParseQuery(string(requestBody))
	if err != nil {
		return params, fmt.Errorf("error parsing body to URL query, data=[%v] error=[%v]", string(requestBody), err)
	}

	// put into map
	for k, v := range paramsPairs {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}
	return
}
