package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// Define a struct to hold the URL mapping
type URLMap struct {
	ShortURL    string
	OriginalURL string
}

// Global map to store the shortened URLs.  Using a simple map for demonstration.
// In a real application, you'd use a database or a more persistent storage.
var urlStore = make(map[string]URLMap)

// Base URL for shortened URLs.  Make this configurable for different environments.
const baseURL = "http://localhost:8080/" //  Include the trailing slash

// Function to generate a short URL from a long URL.
func shortenURL(longURL string) (string, error) {
	// 1. Validate the URL
	if _, err := url.ParseRequestURI(longURL); err != nil {
		return "", fmt.Errorf("invalid URL: %v", err)
	}

	// 2. Check if the URL has already been shortened.
	for _, v := range urlStore {
		if v.OriginalURL == longURL {
			return v.ShortURL, nil // Return the existing short URL
		}
	}

	// 3. Generate a unique short URL.
	hash := md5.Sum([]byte(longURL))
	shortURL := hex.EncodeToString(hash[:])[:8] // Use the first 8 characters of the hash.

	// Ensure the short URL is unique.  Handle collisions.
	counter := 1
	originalShortURL := shortURL
	for _, exists := urlStore[shortURL]; exists; {
		shortURL = fmt.Sprintf("%s%d", originalShortURL, counter) // Append a counter
		counter++
	}

	// 4. Store the mapping in the global map.
	urlStore[shortURL] = URLMap{
		ShortURL:    shortURL,
		OriginalURL: longURL,
	}

	return shortURL, nil
}

// Handler function for the root path ("/").  This serves the HTML form.
func rootHandler(w http.ResponseWriter, r *http.Request) {
	// Serve the HTML file from the templates directory.
	tmpl, err := template.ParseFiles("templates/index.html") // Corrected path
	if err != nil {
		http.Error(w, fmt.Sprintf("Error serving HTML: %v", err), http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, nil) // Pass nil if you don't have data to inject.
	if err != nil {
		http.Error(w, fmt.Sprintf("Error executing template: %v", err), http.StatusInternalServerError)
		return
	}
}

// Handler function for shortening URLs ("/shorten").
func shortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 1. Get the URL from the form data.
	err := r.ParseForm()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing form: %v", err), http.StatusBadRequest)
		return
	}
	longURL := r.FormValue("url")

	if longURL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	// 2. Shorten the URL.
	shortURL, err := shortenURL(longURL)
	if err != nil {
		// Handle the error from shortenURL
		http.Error(w, fmt.Sprintf("Error shortening URL: %v", err), http.StatusBadRequest)
		return
	}

	// 3. Return the shortened URL as JSON.
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"short_url":"%s%s"}`, baseURL, shortURL) // Correctly use baseURL
}

// Handler function to redirect from short URL to original URL.
func redirectHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Get the short URL from the path.
	shortURL := strings.TrimPrefix(r.URL.Path, "/") // Remove the leading slash
	if shortURL == "" || shortURL == "shorten" {    //handle empty and /shorten
		http.NotFound(w, r)
		return
	}

	// 2. Look up the short URL in the storage.
	urlMap, ok := urlStore[shortURL]
	if !ok {
		http.NotFound(w, r) // 404 if short URL not found.
		return
	}

	// 3. Redirect to the original URL.
	http.Redirect(w, r, urlMap.OriginalURL, http.StatusMovedPermanently) // 301 redirect
}

func main() {
	// 1. Set up routing.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			rootHandler(w, r)
		} else {
			redirectHandler(w, r)
		}
	})
	http.HandleFunc("/shorten", shortenHandler) // Handle the URL shortening request

	// 2. Start the server.
	fmt.Println("Server listening on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
