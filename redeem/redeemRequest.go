package redeem

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/Arpitkr/iitk-coin/coins"
	"github.com/Arpitkr/iitk-coin/store"
	_ "github.com/mattn/go-sqlite3"
)

var MyDB *sql.DB
var mutex2 sync.Mutex

type Request struct {
	Roll   int `json:"roll"`
	ItemID int `json:"itemid"`
}

func RedeemRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		ok, claims := coins.VerifyJwt(w, r)
		if !ok {
			return
		}

		request := Request{}
		store := store.Store{}
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			coins.SetError(w, err)
			return
		} else if claims.Roll != request.Roll {
			w.WriteHeader(http.StatusUnauthorized)
			w.Header().Set("Content-type", "json")
			json.NewEncoder(w).Encode("Status Unauthorized")
			return
		}

		rows, err := MyDB.Query("Select * from Store where ItemID = ?", request.ItemID)
		if err != nil {
			coins.SetError(w, err)
			return
		}

		count := 0
		for rows.Next() {
			count++
			rows.Scan(&store.ItemID, &store.ItemName, &store.Cost, &store.Stock)
		}
		if count == 0 || store.Stock <= 0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-type", "json")
			json.NewEncoder(w).Encode("Item Unavailable")
			return
		}

		coin := 0.0
		coins.Mutex.Lock()
		mutex2.Lock()
		rows, err = MyDB.Query("Select Coins from User where Roll = ?", request.Roll)
		if err != nil {
			coins.SetError(w, err)
			mutex2.Unlock()
			coins.Mutex.Unlock()
			return
		}
		for rows.Next() {
			rows.Scan(&coin)
		}
		if coin < store.Cost {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-type", "json")
			json.NewEncoder(w).Encode("Insufficient balance")
			coins.Mutex.Unlock()
			mutex2.Unlock()
			return
		} else {
			_, err := MyDB.Exec("Insert into RedeemStatus(User, ItemID, RequestTime) values(?,?,datetime('now','localtime'))", request.Roll, request.ItemID)
			if err != nil {
				coins.SetError(w, err)
				coins.Mutex.Unlock()
				mutex2.Unlock()
				return
			}
		}
		coins.Mutex.Unlock()
		mutex2.Unlock()
	}
}
