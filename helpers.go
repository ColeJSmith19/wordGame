package main

func getWordWithGuesses(id string) string {
	activeGame := activeGames[id]

	current := activeGame.Current

	if len(activeGame.GuessedLetters) == 0 {
		for i := 0; i < len(activeGame.Word); i++ {
			current += "_"
		}

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
