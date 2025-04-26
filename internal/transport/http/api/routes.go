package api

import (
	"github.com/gorilla/mux"
)

func NewRouter(handler *Handler) *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/login", handler.HandleLogin).Methods("GET")
	router.HandleFunc("/refresh", handler.HandleRefreshToken).Methods("PATCH")
	router.HandleFunc("/register", handler.HandleRegister).Methods("POST")

	return router
}
