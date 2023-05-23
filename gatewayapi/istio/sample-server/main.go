package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	echoMessage := os.Getenv("ECHO_MESSAGE")
	if echoMessage == "" {
		echoMessage = "Hello, World!"
	}

	echoHandler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s", echoMessage)
	}

	rootHandler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s", "OK")
	}

	headersHandler := func(w http.ResponseWriter, r *http.Request) {
		jsonResp, err := json.Marshal(r.Header)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Internal Server Error")
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%s", jsonResp)
	}

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/get", echoHandler)
	http.HandleFunc("/headers", headersHandler)

	log.Println("Startting Server")
	// nil 指定で defaultのマルチプレクサを利用
	log.Fatal(http.ListenAndServe(":8080", nil))
}
