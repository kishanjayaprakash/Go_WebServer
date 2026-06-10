package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// User struct now includes a MongoDB ObjectID field; bson tags control how fields are stored in MongoDB
type User struct {
	ID   int    `bson:"_id" json:"id"`
	Name string `bson:"name" json:"name"`
}

func main() {
	ConnectDB()          // connect to MongoDB on startup
	defer DisconnectDB() // cleanly disconnect from MongoDB when the server shuts down

	mux := http.NewServeMux() //router that matches incoming requests to their respective handlers

	mux.HandleFunc("/", handleRoot) // routes "/" requests to handleRoot function

	mux.HandleFunc("POST /users", createUser)        // routes "POST /users" requests to createUser function
	mux.HandleFunc("GET /users/{id}", getUser)       // routes "GET /users/{id}" requests to getUser function
	mux.HandleFunc("DELETE /users/{id}", deleteUser) // routes "DELETE /users/{id}" requests to deleteUser function

	// Wrap the entire router inside our RateLimitMiddleware
	wrappedMux := RateLimitMiddleware(mux)

	fmt.Println("Server is running on http://localhost:8080") // prints a message to the console indicating that the server is running
	err := http.ListenAndServe(":8080", wrappedMux)           // Pass wrappedMux here instead of mux
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello from root!") //sends get request response with "Hello from root!" message
}

func getUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id")) // parse the string ID from the URL into an integer
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	fmt.Printf("[GET] Request received for ID: %d\n", id) // Diagnostic log

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user User
	err = userCol.FindOne(ctx, bson.M{"_id": id}).Decode(&user) // query MongoDB for the user with the given ID
	if err == mongo.ErrNoDocuments {
		fmt.Printf("[GET] User with ID %d was NOT found in database\n", id) // Diagnostic log
		http.Error(w, "User not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json") // sets the Content-Type header to application/json to indicate that the response is in JSON format

	j, err := json.Marshal(user) // marshals the User struct into JSON format
	if err != nil {
		http.Error(w, "Error encoding user data", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK) // sets the HTTP status code to 200 OK
	w.Write(j)                   // sends the JSON response back to the client
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id")) // parse the string ID from the URL into an integer
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := userCol.DeleteOne(ctx, bson.M{"_id": id}) // delete the user from MongoDB with the given ID
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if result.DeletedCount == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent) // sets the HTTP status code to 204 No Content to indicate that the user has been successfully deleted
}

// runs when POST /users is called to create a new user in database
func createUser(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user) // decodes the JSON request body into a User struct
	if err != nil {
		fmt.Println("[POST] Failed to decode request body:", err) // Diagnostic log
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if user.Name == "" {
		fmt.Println("[POST] Rejected request: 'name' field is empty") // Diagnostic log
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	id, err := getNextID() // generate the next auto-incremented integer ID from the counters collection
	if err != nil {
		http.Error(w, "Failed to generate user ID", http.StatusInternalServerError)
		return
	}
	user.ID = id // assign the generated ID to the user before inserting into MongoDB

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = userCol.InsertOne(ctx, user) // inserts the new user into MongoDB and gets back the generated ID
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	fmt.Printf("[POST] Successfully saved user '%s' with ID: %d\n", user.Name, user.ID) // Diagnostic log

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // sets the HTTP status code to 201 Created to indicate the user was successfully created
	json.NewEncoder(w).Encode(user)   // sends the newly created user (including their MongoDB ID) back to the client
}