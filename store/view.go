package store

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/Arpitkr/iitk-coin/coins"
	_ "github.com/mattn/go-sqlite3"
)

var MyDB *sql.DB

func SetError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Content-type", "json")
	json.NewEncoder(w).Encode("Internal server error")
	log.Print(err)
}

type Store struct {
	ItemID   int     `json:"itemid"`
	ItemName string  `json:"name"`
	Cost     float64 `json:"cost"`
	Stock    int     `json:"stock"`
}

func ViewStore(w http.ResponseWriter, r *http.Request) {
	ok, _ := coins.VerifyJwt(w, r)
	if !ok {
		return
	}

	rows, err := MyDB.Query("Select * from Store")
	if err != nil {
		SetError(w, err)
		return
	}

	store := Store{}

	for rows.Next() {
		rows.Scan(&store.ItemID, &store.ItemName, &store.Cost, &store.Stock)
		json.NewEncoder(w).Encode(store)
	}
}
