package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	helloHandler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, World!")
	}
	rootHandler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Root Handler!")
	}
	http.HandleFunc("/hello", helloHandler)
	http.HandleFunc("/", rootHandler)

	log.Println("Startting Server")
	// nil 指定で defaultのマルチプレクサを利用
	log.Fatal(http.ListenAndServe(":8080", nil))
}
