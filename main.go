package main

import (
	"fmt"
	"log"
	"net/http"

	"go-paymob/handlers"
)

func main() {
	http.HandleFunc("/pay", handlers.PayHandler)

	fmt.Println("🚀 Server running on http://localhost:8081")
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatal(err)
	}
}
