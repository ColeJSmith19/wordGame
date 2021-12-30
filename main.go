package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
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

func getWordWithGuesses(id string) string {
	activeGame := activeGames[id]

	current := activeGame.Current

	if len(activeGame.GuessedLetters) == 0 {
		for i := 0; i < len(activeGame.Word); i++ {
			current += "_"
		}

		activeGame.Current = current
		return current
	}

	finalWord := ""
	for _, c := range activeGame.Word {
		letterFound := false
		for _, gl := range activeGame.GuessedLetters {
			if string(c) == gl {
				letterFound = true
				continue
			}
		}
		if !letterFound {
			finalWord += "_"
		} else {
			finalWord += string(c)
		}
	}

	return finalWord
}

func newGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Must be a 'GET' request", http.StatusBadRequest)
		return
	}

	id, err := generateIdentifier()
	if err != nil {
		http.Error(w, "Unable to generate id", http.StatusInternalServerError)
		return
	}
	gameWord := gameWords[rand.Intn(len(gameWords))]

	game := activeGame{ID: id, Current: "", GuessesRemaining: 6, Word: gameWord, GuessedLetters: []string{}}
	activeGames[id] = game
	rsp := playerGameRsp{ID: id, Current: getWordWithGuesses(id), GuessesRemaining: 6}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rsp)
}

func makeGuess(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Must be a 'POST' request", http.StatusBadRequest)
		return
	}

	var guess playerGuess

	err := json.NewDecoder(r.Body).Decode(&guess)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if guess.ID == "" {
		http.Error(w, "'id' is required", http.StatusBadRequest)
		return
	}

	if guess.Guess == "" {
		http.Error(w, "'guess' is required", http.StatusBadRequest)
		return
	}

	if _, ok := activeGames[guess.ID]; !ok {
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}

	activeGame := activeGames[guess.ID]

	match, _ := regexp.MatchString("[A-Z]", guess.Guess)
	if !match {
		http.Error(w, "'guess' must be a single, captial letter [A-Z]", http.StatusNotFound)
		return
	}

	activeGame.GuessedLetters = append(activeGame.GuessedLetters, guess.Guess)
	activeGames[guess.ID] = activeGame

	if !strings.Contains(activeGame.Word, guess.Guess) {
		activeGame.GuessesRemaining--
	}

	if activeGame.GuessesRemaining == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(fmt.Sprintf("That was your last guess. The word was %s, you guessed the following letters %v", activeGame.Word, activeGame.GuessedLetters))
		delete(activeGames, guess.ID)
		return
	}

	activeGame.Current = getWordWithGuesses(activeGame.ID)

	if !strings.Contains(activeGame.Current, "_") {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(fmt.Sprintf("You win! The word was %s, you guessed the following letters %v", activeGame.Word, activeGame.GuessedLetters))
		delete(activeGames, guess.ID)
		return
	}

	activeGames[guess.ID] = activeGame

	rsp := playerGameRsp{ID: activeGame.ID, Current: activeGame.Current, GuessesRemaining: activeGame.GuessesRemaining}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rsp)
}

func viewGames(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(activeGames)
}
