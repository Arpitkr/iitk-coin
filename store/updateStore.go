package store

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Arpitkr/iitk-coin/coins"
	_ "github.com/mattn/go-sqlite3"
)

type Stock struct {
	ItemID int `json:"itemid"`
	Change int `json:"change"`
}

func Verify(w http.ResponseWriter, r *http.Request) bool {
	ok, claims := coins.VerifyJwt(w, r)
	if !ok {
		return false
	}

	if !claims.IsAdmin {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-type", "json")
		json.NewEncoder(w).Encode("Status Unauthorized")
		return false
	}
	return true
}

func IncreaseStock(w http.ResponseWriter, ItemID int, change int) bool {
	rows, err := MyDB.Query("Select * from Store where ItemID = ?", ItemID)
	if err != nil {
		SetError(w, err)
		return false
	}
	count := 0
	for rows.Next() {
		count++
	}

	if count == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-type", "json")
		json.NewEncoder(w).Encode(fmt.Sprint("Invalid ItemID ", ItemID))
		return false
	}

	res, err := MyDB.Exec("Update Store set Stock = Stock + ? where ItemID = ? and Stock + ? >= 0", change, ItemID, change)
	if err != nil {
		SetError(w, err)
		return false
	} else if check, _ := res.RowsAffected(); check == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-type", "json")
		json.NewEncoder(w).Encode("Invalid Change")
		return false
	} else {
		return true
	}
}

func AddStore(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {

		if check := Verify(w, r); !check {
			return
		}

		store := Store{}
		err := json.NewDecoder(r.Body).Decode(&store)
		if err != nil {
			SetError(w, err)
			return
		}

		coins.Mutex.Lock()
		tx, err := MyDB.Begin()
		if err != nil {
			SetError(w, err)
			tx.Rollback()
			coins.Mutex.Unlock()
			return
		}

		_, err = tx.Exec("Insert into Store values(?,?,?,?)", store.ItemID, store.ItemName, store.Cost, store.Stock)
		if err != nil {
			SetError(w, err)
			tx.Rollback()
			coins.Mutex.Unlock()
			return
		} else {
			tx.Commit()
			coins.Mutex.Unlock()
		}
	}
}

func DeleteStore(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {

		if check := Verify(w, r); !check {
			return
		}

		store := Store{}
		err := json.NewDecoder(r.Body).Decode(&store)
		if err != nil {
			SetError(w, err)
			return
		}

		coins.Mutex.Lock()
		tx, err := MyDB.Begin()
		if err != nil {
			SetError(w, err)
			tx.Rollback()
			coins.Mutex.Unlock()
			return
		}

		_, err = tx.Exec("Delete from Store where ItemID = ?", store.ItemID)
		if err != nil {
			SetError(w, err)
			tx.Rollback()
			coins.Mutex.Unlock()
			return
		} else {
			tx.Commit()
			coins.Mutex.Unlock()
		}
	}
}

//Receives a post request containing ItemID and change in number of items.
func ChangeStock(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if check := Verify(w, r); !check {
			return
		}

		stock := Stock{}
		err := json.NewDecoder(r.Body).Decode(&stock)
		if err != nil {
			SetError(w, err)
			return
		}

		coins.Mutex.Lock()
		IncreaseStock(w, stock.ItemID, stock.Change)
		coins.Mutex.Unlock()
	}
}
