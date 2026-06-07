package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
)

type User struct {
	Name string `json:"name"`

}

var usercache = make(map[int]User) // in-memory database, key is user id and value is User struct

var cachemutex sync.RWMutex // mutex to protect access to the usercache map

func main() {
	mux := http.NewServeMux() //router that matches incoming requests to their respective handlers

	mux.HandleFunc("/", handleRoot) // routes "/" requests to handleRoot function

    mux.HandleFunc( "POST /users", createUser) // routes "POST /users" requests to createUser function
    mux.HandleFunc("GET /users/{id}", getUser) // routes "GET /users/{id}" requests to getUser function

	fmt.Println("Server is running on http://localhost:8080") // prints a message to the console indicating that the server is running
	err := http.ListenAndServe(":8080", mux)  
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello from root!") //sends get request response with "Hello from root!" message
}

func getUser(w http.ResponseWriter, r *http.Request) {
	id, err:= strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	cachemutex.RLock() // locks the mutex for reading to allow multiple goroutines to read from the usercache map concurrently
	user, ok := usercache[id]
	cachemutex.RUnlock() // unlocks the mutex after reading from the map
	if !ok {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

    w.Header().Set("Content-Type", "application/json") // sets the Content-Type header to application/json to indicate that the response is in JSON format

	j, err := json.Marshal(user) // marshals the User struct into JSON format
	if err != nil {
		http.Error(w, "Error encoding user data", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK) // sets the HTTP status code to 200 OK
	w.Write(j) // sends the JSON response back to the client
	
}

// runs when POST /users is called
func createUser(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user) // decodes the JSON request body into a User struct
	if err != nil {
		http.Error(w,err.Error(), http.StatusBadRequest)
		return
	}
	if user.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}
	cachemutex.Lock() // locks the mutex to ensure that only one goroutine can access the usercache map at a time
	usercache[len(usercache)+1] = user // adds the user to the usercache map with a new id
	cachemutex.Unlock() // unlocks the mutex after the user has been added to the map
	w.WriteHeader(http.StatusNoContent) 
}
