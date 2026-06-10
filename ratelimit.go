package main

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

// Visitor struct to track the number of requests and the last seen time for each IP address
type Visitor struct {
	count    int
	lastSeen time.Time
}

var (
	visitors      = make(map[string]*Visitor) // Map to store visitors with their IP as the key
	visitorsMutex sync.Mutex                  // Mutex to protect access to the visitors mapMutex
)

func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr) // Extract the client's IP address from the request
		if err != nil {
			ip = r.RemoteAddr // Fallback to the full RemoteAddr if splitting fails
		}

		now := time.Now() // Define the 'now' timestamp variable so it can be used below

		visitorsMutex.Lock() // changed visitorMutex to visitorsMutex
		v, exists := visitors[ip]
		if !exists {
			visitors[ip] = &Visitor{count: 1, lastSeen: now} // If the visitor is new, create a new Visitor entry
			visitorsMutex.Unlock()                           // changed visitorMutex to visitorsMutex
			next.ServeHTTP(w, r)
			return
		}
		if now.Sub(v.lastSeen) > time.Minute { // if the last request was more than 1 minute ago, reset the counter
			v.count = 1
			v.lastSeen = now
			visitorsMutex.Unlock()
			next.ServeHTTP(w, r)
			return
		}
		v.count++
		v.lastSeen = now
		if v.count > 15 {
			visitorsMutex.Unlock()                                                                    // Fixed typo: changed visitorMutex to visitorsMutex
			fmt.Printf("[RateLimit] IP %s has exceeded the request limit. Count: %d\n", ip, v.count) // Diagnostic log
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}
		visitorsMutex.Unlock() // Fixed typo: changed visitorMutex to visitorsMutex
		next.ServeHTTP(w, r)   // Call the next handler in the chain
	})
}