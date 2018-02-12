package main

import (
	"fmt"

	"github.com/muesli/regommend"
)

type Recommender struct {
	filmsTable *regommend.RegommendTable
}

// Init DB config & refresh recommender.
func NewRecommender() (re Recommender) {
	re.filmsTable = regommend.Table("films")
	re.refresh()
	return re
}

// Pull film data from SQL DB and parse into regommender object.
func (re *Recommender) refresh() (err error) {
	re.filmsTable.Flush()

	watchedFilms := make(map[interface{}]float64)
	watchedFilms["film_id"] = 4.0

	re.filmsTable.Add(12345, watchedFilms)

	return nil
}

// Produce a set of recommendations for a user of a set size.
func (re *Recommender) recommend(userID string, num int) (response string) {
	// get recommendations
	recs, err := re.filmsTable.Recommend(userID)
	if err != nil {
		return fmt.Sprintf("could not generate recommendations for user, ID:[%v], error:[%v]", userID, err)
	}

	if len(recs) < num {
		num = len(recs)
	}
	for result := 0; result < num; result++ {
		fmt.Printf("[%v]: Recommending %v with score %v\n", result, recs[result].Key, recs[result].Distance)
	}

	return ""
}
