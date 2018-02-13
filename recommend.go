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

// Pull fresh film data from SQL DB and parse into regommend object.
func (re *Recommender) refresh() (err error) {
	re.filmsTable.Flush()

	req, err := dbInstance.connect()
	if err != nil {
		return err
	}

	watchedData := req.GetAllWatchedListData()
	if err != nil {
		return err
	}

	// iterate over films
	for userID, filmRecord := range *watchedData {
		re.filmsTable.Add(userID, filmRecord)
	}


	return nil
}

// Produce a set of recommendations for a user of a set size.
func (re *Recommender) recommend(userID int, numResults int) (response map[int]float64, err error) {
	// pull fresh data from DB
	if err = re.refresh(); err != nil {
		return
	}

	// generate recommendations
	recs, err := re.filmsTable.Recommend(userID)
	if err != nil {
		return
	}

	// trim result set to requested length
	if len(recs) < numResults || numResults == 0 {
		numResults = len(recs)
	}

	// parse for response
	response = make(map[int]float64)
	for result := 0; result < numResults; result++ {
		filmID := recs[result].Key.(int)
		response[filmID] = recs[result].Distance
		fmt.Printf("[%v]: Recommending %v with score %v\n", result, recs[result].Key, recs[result].Distance)
	}

	return response, nil
}
