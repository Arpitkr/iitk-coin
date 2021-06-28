package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/Arpitkr/iitk-coin/auth"
	"github.com/Arpitkr/iitk-coin/coins"
	"github.com/Arpitkr/iitk-coin/data"
	_ "github.com/mattn/go-sqlite3"
)

var MyDB *sql.DB
var TransferDB *sql.DB
var AwardDB *sql.DB

func handleRequest() {
	//Adding handler functions for different end points
	http.HandleFunc("/signup", auth.Signup)
	http.HandleFunc("/login", auth.Login)
	http.HandleFunc("/secretpage", data.Secret)
	http.HandleFunc("/rewardCoins", coins.AwardCoin)
	http.HandleFunc("/transferCoins", coins.TransferCoin)
	http.HandleFunc("/getCoins", coins.GetCoins)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode("Thank you for accessing this page. For signup access /signup. For login access /login.")
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func main() {
	MyDB, err := sql.Open("sqlite3", "../Info.db")
	if err != nil {
		log.Fatal(err)
	}
	auth.MyDB = MyDB
	coins.MyDB = MyDB
	_, err = MyDB.Exec("Create table If not exists TransferHistory(Sender Integer, Receiver Int, SentAmount Float, ReceivedAmount Float, Tax Float, Timestamp varchar(20))")
	if err != nil {
		log.Fatal(err)
	}
	_, err = MyDB.Exec("Create table If not exists RewardHistory(Receiver Int, ReceivedAmount Float, Timestamp varchar(20))")
	if err != nil {
		log.Fatal(err)
	}
	handleRequest()
}
