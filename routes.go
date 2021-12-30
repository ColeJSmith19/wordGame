package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
)

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
