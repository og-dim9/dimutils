package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	port := 4321
	// Check if the "--slow" command-line argument is provided
	slow := false
	for _, arg := range os.Args {
		if arg == "--slow" {
			slow = true
			break
		}
	}
	// Create a custom logger
	logger := log.New(os.Stdout, "", log.LstdFlags)

	// Middleware function to log each page access
	logMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Printf("%s", r.URL.Path)
			if slow {
				time.Sleep(2 * time.Second) // Simulate a slow server
			}
			next.ServeHTTP(w, r)
		})
	}
	// Use the middleware
	http.Handle("/", logMiddleware(http.FileServer(http.Dir("."))))

	// Start the server
	addr := fmt.Sprintf(":%d", port)
	log.Printf("Server listening on port %d", port)

	log.Fatal(http.ListenAndServe(addr, nil))
}
