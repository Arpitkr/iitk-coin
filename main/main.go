package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Arpitkr/iitk-coin/auth"
	"github.com/Arpitkr/iitk-coin/data"
)

func handleRequest() {
	//Adding handler functions for different end points
	http.HandleFunc("/signup", auth.Signup)
	http.HandleFunc("/login", auth.Login)
	http.HandleFunc("/secretpage", data.Secret)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode("Thank you for accessing this page. For signup access /signup. For login access /login.")
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func main() {
	handleRequest()
}
