package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/Arpitkr/iitk-coin/auth"
	"github.com/Arpitkr/iitk-coin/coins"
	"github.com/Arpitkr/iitk-coin/data"
	"github.com/Arpitkr/iitk-coin/redeem"
	"github.com/Arpitkr/iitk-coin/store"
	_ "github.com/mattn/go-sqlite3"
)

var MyDB *sql.DB

func handleRequest() {
	// Adding handler functions for different end points
	http.HandleFunc("/signup", auth.Signup)
	http.HandleFunc("/login", auth.Login)
	http.HandleFunc("/secretpage", data.Secret)
	http.HandleFunc("/rewardCoins", coins.AwardCoin)
	http.HandleFunc("/transferCoins", coins.TransferCoin)
	http.HandleFunc("/getCoins", coins.GetCoins)
	http.HandleFunc("/store/view", store.ViewStore)
	http.HandleFunc("/store/update/add", store.AddStore)
	http.HandleFunc("/store/update/delete", store.DeleteStore)
	http.HandleFunc("/store/update/stock", store.ChangeStock)
	http.HandleFunc("/redeem/view", redeem.ViewRedeem)
	http.HandleFunc("/redeem/request", redeem.RedeemRequest)
	http.HandleFunc("/redeem/action", redeem.RedeemAction)
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
	store.MyDB = MyDB
	redeem.MyDB = MyDB
	_, err = MyDB.Exec("Create table If not exists TransferHistory(Sender Integer, Receiver Integer, SentAmount Float, ReceivedAmount Float, Tax Float, Timestamp varchar(20))")
	if err != nil {
		log.Fatal(err)
	}
	_, err = MyDB.Exec("Create table If not exists RewardHistory(Receiver Integer, ReceivedAmount Float, Timestamp varchar(20))")
	if err != nil {
		log.Fatal(err)
	}
	_, err = MyDB.Exec("Create table If not exists RedeemStatus(RedeemID Integer primary key autoincrement,User Integer, ItemID integer, Status Varchar(20) default 'PENDING', RequestTime varchar(20), ResponseTime varchar(20))")
	if err != nil {
		log.Fatal(err)
	}
	_, err = MyDB.Exec("Create table If not exists Store(ItemID integer primary key, ItemName varchar(20), Cost float,  Stock Integer default 0)")
	if err != nil {
		log.Fatal(err)
	}
	handleRequest()
}
