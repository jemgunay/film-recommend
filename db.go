package main

import (
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"strings"
)

type DBWrapper struct {
	sqlUser   string
	sqlPass   string
	sqlDBName string
}

func NewDatabase() (w DBWrapper) {
	w.sqlUser = "root"
	w.sqlPass = ""
	w.sqlDBName = "film_service"
	// set gorm default capitalised table names
	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		return strings.Title(defaultTableName[:len(defaultTableName)-1])
	}

	w.perform("GetUserByName", "rob")

	return w
}

// Connect to the DB, perform an operation, and close the connection.
func (w *DBWrapper) perform(operation string, params ...string) (result interface{}, err error) {
	// connect to DB
	db, err := gorm.Open("mysql", fmt.Sprintf("%v:%v@/%v?charset=utf8&parseTime=True&loc=Local", w.sqlUser, w.sqlPass, w.sqlDBName))
	if err != nil {
		return nil, err
	}
	defer db.Close()

	switch operation {
	case "GetUserByName":
		result, err = w.GetUserByName(db, params[0])

	case "GetWatchedByUserID":
		result, err = w.GetWatchedByUserID(db, params[0])
	}
	return
}

// Model for User table.
type User struct {
	ID   int   `gorm:"column:user_id;primary_key;AUTO_INCREMENT"`
	Name string `gorm:"column:name"`
}

// Get the user ID for the corresponding user name.
func (w *DBWrapper) GetUserByName(db *gorm.DB, name string) (response *User, err error) {
	userResult := new(User)
	db.Where("name = ?", name).First(userResult)

	fmt.Println(userResult)

	return userResult, nil
}

// Model for Watched table.
type Watched struct {
	ID   int   `gorm:"column:watched_id;primary_key;AUTO_INCREMENT"`
	UserID int `gorm:"column:user_id"`
	FilmID int `gorm:"column:film_id"`
	Rating int `gorm:"column:rating"`
}

// Get the user ID for the corresponding user name.
func (w *DBWrapper) GetWatchedByUserID(db *gorm.DB, userID string) (response *[]Watched, err error) {
	var watchedResults *[]Watched
	db.Where("user_id = ?", userID).Find(watchedResults)

	fmt.Println(watchedResults)

	return watchedResults, nil
}