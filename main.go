package main

import (
	"fmt"
	"net/http"
)

func main() {
	mux := http.NewServeMux() //router that matches incoming requests to their respective handlers

	mux.HandleFunc("/", handleRoot) // routes "/" requests to handleRoot function
    
	fmt.Println("Server is running on http://localhost:8080") // prints a message to the console indicating that the server is running
	http.ListenAndServe(":8080", mux) // starts the server and listens on port 8080 and sends requests to mux
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello from root!")
}