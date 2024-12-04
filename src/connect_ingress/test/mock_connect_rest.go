package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	case "GET":
		fmt.Fprintf(w, "GET request\n")
	case "POST":
		fmt.Fprintf(w, "POST request\n")
	case "PUT":
		fmt.Fprintf(w, "PUT request\n")
	case "DELETE":
		fmt.Fprintf(w, "DELETE request\n")
	default:
		fmt.Fprintf(w, "Unsupported request\n")

}