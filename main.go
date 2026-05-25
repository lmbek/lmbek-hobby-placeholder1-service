package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "User Service is Healthy")
	})

	fmt.Println("User Service starting on :8082...")
	if err := http.ListenAndServe(":8082", nil); err != nil {
		fmt.Printf("Error starting service: %v\n", err)
	}
}
