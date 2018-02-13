package main

import (
	"fmt"
	"strings"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// Contains DB connection settings.
type DBInstance struct {
	sqlUser   string
	sqlPass   string
	sqlDBName string
}

// Create instance and set DB settings.
func NewDBInstance() (w DBInstance) {
	w.sqlUser = "root"
	w.sqlPass = ""
	w.sqlDBName = "film_service"

	// set gorm default capitalised table names and remove plural s
	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		return strings.Title(defaultTableName[:len(defaultTableName)-1])
	}

	return w
}

// Container to persist db access.
type DBRequest struct {
	db *gorm.DB
}

// Connect to the DB.
func (w *DBInstance) connect() (req DBRequest, err error) {
	req.db, err = gorm.Open("mysql", fmt.Sprintf("%v:%v@/%v?charset=utf8&parseTime=True&loc=Local", w.sqlUser, w.sqlPass, w.sqlDBName))
	return
}

// Model for User table.
type User struct {
	ID   int    `gorm:"column:user_id;primary_key;AUTO_INCREMENT"`
	Name string `gorm:"column:name"`
}

// Get the user ID for the corresponding user name.
func (r *DBRequest) GetUserByName(name string) (response *User, err error) {
	defer r.db.Close()

	var userResult *User
	r.db.Where("name = ?", name).First(userResult)

	return userResult, nil
}

// Model for Watched table.
type Watched struct {
	ID     int `gorm:"column:watched_id;primary_key;AUTO_INCREMENT"`
	UserID int `gorm:"column:user_id"`
	FilmID int `gorm:"column:film_id"`
	Rating int `gorm:"column:rating"`
}

// Get the user ID for the corresponding user name.
func (r *DBRequest) GetAllWatchedListData() (response *map[int]map[interface{}]float64) {
	defer r.db.Close()

	var watchedResults []Watched
	r.db.Find(&watchedResults)

	watchedLists := make(map[int]map[interface{}]float64)

	for _, record := range watchedResults {
		// check if user has been found yet
		if _, ok := watchedLists[record.UserID]; !ok {
			watchedLists[record.UserID] = make(map[interface{}]float64)
		}

		// add film & rating record to user
		m := watchedLists[record.UserID]
		m[record.FilmID] = float64(record.Rating)
		watchedLists[record.UserID] = m
	}

	return &watchedLists
}

// Get all records from the User table.
func (r *DBRequest) GetUsers(userID string) (response *[]User) {
	defer r.db.Close()

	var usersResults []User
	r.db.Find(&usersResults)

	return &usersResults
}

// Get the user ID for the corresponding user name.
func (r *DBRequest) GetWatchedByUserID(userID string) (response *[]Watched) {
	defer r.db.Close()

	var watchedResults []Watched
	r.db.Where("user_id = ?", userID).Find(&watchedResults)

	return &watchedResults
}

// Add a film to a user's watched list.
func (r *DBRequest) AddFilmToWatchedList(userID int, filmID int, rating int) (err error) {
	defer r.db.Close()

	newWatched := Watched{UserID: userID, FilmID: filmID, Rating: rating}

	if ok := r.db.NewRecord(newWatched); !ok {
		return fmt.Errorf("film already found in watched list")
	}

	r.db.Create(&newWatched)

	return nil
}
