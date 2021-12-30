package main

import (
	"log"
	"math/rand"
	"net/http"
	"time"
)

const (
	serverAddress = "localhost:1337"
)

type playerGameRsp struct {
	ID               string `json:"id"`
	Current          string `json:"current"`
	GuessesRemaining int    `json:"guesses_remaining"`
}

type activeGame struct {
	ID               string
	Current          string
	GuessesRemaining int
	Word             string
	GuessedLetters   []string
}

type playerGuess struct {
	ID    string `json:"id"`
	Guess string `json:"guess"`
}

var gameWords []string

// postgreSQL connetion to store this across sessions?
var activeGames map[string]activeGame

func main() {
	rand.Seed(time.Now().UnixNano())

	words, err := loadWords("words.txt")
	if err != nil {
		log.Fatal(err)
	}

	// TODO use words in your implementation
	gameWords = words

	activeGames = make(map[string]activeGame)

	// create, store and start a new game
	http.HandleFunc("/new", newGame)

	// sanity check endpoint
	http.HandleFunc("/view_games", viewGames)

	// make a guess against an existing game
	http.HandleFunc("/guess", makeGuess)

	log.Printf("Starting server on http://%s", serverAddress)
	if err := http.ListenAndServe(serverAddress, nil); err != nil {
		log.Fatal(err)
	}
}
