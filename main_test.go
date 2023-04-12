package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetRandomDadJoke(t *testing.T) {
	// Set up a test server to handle requests and return a predefined joke
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"joke":"Why did the tomato turn red? Because it saw the salad dressing!"}`)
	}))
	defer ts.Close()

	// Call getRandomDadJoke with the test server URL and check that it returns the expected joke
	joke, err := getRandomDadJoke(ts.URL)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	expectedJoke := "Why did the tomato turn red? Because it saw the salad dressing!"
	if joke != expectedJoke {
		t.Errorf("Expected joke '%s', got '%s'", expectedJoke, joke)
	}

}
