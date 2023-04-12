package main

import (
	"math/rand"
	"time"
	"fmt"
	"log"
	"net/http"
	"io"
	"encoding/json"
)
	

type Joke struct {
	Joke string `json:"joke"`
}

const url = "https://icanhazdadjoke.com/"

func getRandomDadJoke(url string) (string, error) {
	rand.Seed(time.Now().UnixNano())

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error getting response: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	var joke Joke
	err = json.Unmarshal(body, &joke)
	if err != nil {
		return "", fmt.Errorf("error decoding response: %v", err)
	}

	return joke.Joke, nil
}

func main() {
	joke, err := getRandomDadJoke(url)
	if err != nil {
		log.Fatalln(err)
		return
	}

	fmt.Println(joke)
}

