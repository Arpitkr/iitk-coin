package redeem

import (
	"encoding/json"
	"net/http"

	"github.com/Arpitkr/iitk-coin/coins"
	"github.com/Arpitkr/iitk-coin/store"
)

type Action struct {
	RedeemID int  `json:"redeemid"`
	Approved bool `json:"approved"`
}

func RedeemAction(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		ok := store.Verify(w, r)
		if !ok {
			return
		}

		action := Action{}
		item := store.Store{}
		err := json.NewDecoder(r.Body).Decode(&action)
		if err != nil {
			coins.SetError(w, err)
			return
		}

		if action.Approved {
			mutex2.Lock()
			coins.Mutex.Lock()
			tx, err := MyDB.Begin()
			if err != nil {
				coins.SetError(w, err)
				mutex2.Unlock()
				coins.Mutex.Unlock()
				return
			}

			rows, err := tx.Query("Select User,ItemID from RedeemStatus where RedeemID = ? and status = 'PENDING'", action.RedeemID)
			if err != nil {
				coins.SetError(w, err)
				tx.Rollback()
				mutex2.Unlock()
				coins.Mutex.Unlock()
				return
			}

			red := Redeem{}
			count := 0
			for rows.Next() {
				count++
				rows.Scan(&red.Roll, &red.ItemID)
			}
			if count == 0 {
				w.WriteHeader(http.StatusBadRequest)
				w.Header().Set("Content-type", "json")
				json.NewEncoder(w).Encode("Invalid RedeemID")
				tx.Rollback()
				mutex2.Unlock()
				coins.Mutex.Unlock()
				return
			}

			rows, err = MyDB.Query("Select * from Store where ItemID = ?", red.ItemID)
			if err != nil {
				coins.SetError(w, err)
				tx.Rollback()
				mutex2.Unlock()
				coins.Mutex.Unlock()
				return
			}

			count = 0
			for rows.Next() {
				count++
				rows.Scan(&item.ItemID, &item.ItemName, &item.Cost, &item.Stock)
			}
			if count == 0 || item.Stock <= 0 {
				w.WriteHeader(http.StatusBadRequest)
				w.Header().Set("Content-type", "json")
				json.NewEncoder(w).Encode("Item Unavailable")
				tx.Rollback()
				mutex2.Unlock()
				coins.Mutex.Unlock()
				return
			}

			rows, err = tx.Query("Select Coins from User where Roll = ?", red.Roll)
			if err != nil {
				coins.SetError(w, err)
				tx.Rollback()
				mutex2.Unlock()
				coins.Mutex.Unlock()
				return
			}
			coin := 0.0
			for rows.Next() {
				rows.Scan(&coin)
			}

			if coin < item.Cost {
				w.WriteHeader(http.StatusBadRequest)
				w.Header().Set("Content-type", "json")
				json.NewEncoder(w).Encode("Insufficient balance")
				tx.Rollback()
				coins.Mutex.Unlock()
				mutex2.Unlock()
				return
			}

			_, err = tx.Exec("Update RedeemStatus set Status = 'Approved' , ResponseTime = datetime('now','localtime') where RedeemID = ? ", action.RedeemID)
			if err != nil {
				coins.SetError(w, err)
				tx.Rollback()
				mutex2.Unlock()
				coins.Mutex.Unlock()
				return
			}
			_, err = tx.Exec("Update User set Coins = Coins - ? where Roll = ?", item.Cost, red.Roll)
			if err != nil {
				coins.SetError(w, err)
				tx.Rollback()
				mutex2.Unlock()
				coins.Mutex.Unlock()
				return
			}

			_, err = tx.Exec("Update Store set Stock = Stock + ? where ItemID = ? and Stock + ? >= 0", -1, red.ItemID, -1)
			if err != nil {
				coins.SetError(w, err)
				tx.Rollback()
				mutex2.Unlock()
				coins.Mutex.Unlock()
				return
			}

			json.NewEncoder(w).Encode("Redeem Successfull")
			tx.Commit()
			coins.Mutex.Unlock()
			mutex2.Unlock()
		} else {
			mutex2.Lock()
			coins.Mutex.Lock()
			tx, err := MyDB.Begin()
			_, err = tx.Exec("Update RedeemStatus set Status = 'Rejected' , ResponseTime = datetime('now','localtime') where RedeemID = ? ", action.RedeemID)
			if err != nil {
				coins.SetError(w, err)
				tx.Rollback()
				mutex2.Unlock()
				coins.Mutex.Unlock()
				return
			}
			tx.Commit()
			coins.Mutex.Unlock()
			mutex2.Unlock()
		}
	}
}
