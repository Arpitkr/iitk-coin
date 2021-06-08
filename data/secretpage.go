package data

import (
	"encoding/json"
	"net/http"

	"github.com/Arpitkr/iitk-coin/auth"
	"github.com/dgrijalva/jwt-go"
)

//Check validity of jwt
func isAuthorized(tokenString string) (bool, auth.Claims) {

	claims := auth.Claims{}
	_, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(auth.Key), nil
	})
	if err != nil {
		return false, claims
	}
	return true, claims
}

func secretMessage() string {
	return "You are eligible to read the secret message"
}

func Secret(w http.ResponseWriter, r *http.Request) {

	//Check if JWT is set
	if _, ok := r.Header["Authorization"]; !ok || r.Header["Authorization"] == nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode("JWT not set")
		return
	}

	//JWT is stored in Authorization Header of http request
	tokenString := r.Header["Authorization"][0]
	check, claims := isAuthorized(tokenString)

	if !check {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode("Invalid JWT")
		return
	}

	//if role is admin, display secret message
	if claims.Role == "admin" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode("Authorized : " + secretMessage())
		return
	} else {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode("Unauthorized to view the secret message\n")
		return
	}
}
