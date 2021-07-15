package redeem

import (
	"encoding/json"
	"net/http"

	"github.com/Arpitkr/iitk-coin/coins"
	"github.com/Arpitkr/iitk-coin/store"
)

type Redeem struct {
	RedeemID    int    `json:"redeemid"`
	Roll        int    `json:"roll"`
	ItemID      int    `json:"itemid"`
	Status      string `json:"status"`
	RequestTime string `json:"requesttime"`
	RedeemTime  string `json:"redeemtime"`
}

func ViewRedeem(w http.ResponseWriter, r *http.Request) {
	ok := store.Verify(w, r)
	if !ok {
		return
	}

	redeem := Redeem{}
	mutex2.Lock()
	rows, err := MyDB.Query("Select * from RedeemStatus where Status = 'PENDING'")
	if err != nil {
		coins.SetError(w, err)
		mutex2.Unlock()
		return
	}
	for rows.Next() {
		rows.Scan(&redeem.RedeemID, &redeem.Roll, &redeem.ItemID, &redeem.Status, &redeem.RequestTime, &redeem.RedeemTime)
		json.NewEncoder(w).Encode(redeem)
	}
	mutex2.Unlock()
}
