package main

import (
	"fmt"
	"music-streaming/file-storage/handlers"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.Handle("/music", &handlers.MusicHandler{})
	mux.Handle("/profile", &handlers.ProfileHandler{})

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		fmt.Println(err.Error())
	}
}
