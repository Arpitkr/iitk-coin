package coins

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

type roll struct {
	R int `json:"roll"`
}

func GetCoins(w http.ResponseWriter, r *http.Request) {
	var rs roll
	err := json.NewDecoder(r.Body).Decode(&rs)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-type", "json")
		json.NewEncoder(w).Encode("Error while parsing credentials")
		log.Print(err)
		return
	}

	//Check if Roll number exists in database
	if !CheckRoll(rs.R, w) {
		return
	}

	Mutex.Lock()

	rows, err := MyDB.Query("Select Coins from User where Roll = ?", rs.R)
	if err != nil {
		SetError(w, err)
		return
	}
	coins := 0
	for rows.Next() {
		rows.Scan(&coins)
	}
	Mutex.Unlock()
	json.NewEncoder(w).Encode(fmt.Sprint("Coins :", coins))
}
