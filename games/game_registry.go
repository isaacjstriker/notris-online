package games

import (
    "math/rand"
    "time"
)

// GameRegistry manages all available games
type GameRegistry struct {
    games []Game
    rng   *rand.Rand
}

// NewGameRegistry creates a new game registry
func NewGameRegistry() *GameRegistry {
    return &GameRegistry{
        games: make([]Game, 0),
        rng:   rand.New(rand.NewSource(time.Now().UnixNano())),
    }
}

// RegisterGame adds a game to the registry
func (gr *GameRegistry) RegisterGame(game Game) {
    gr.games = append(gr.games, game)
}

// GetAllGames returns all registered games
func (gr *GameRegistry) GetAllGames() []Game {
    available := make([]Game, 0)
    for _, game := range gr.games {
        if game.IsAvailable() {
            available = append(available, game)
        }
    }
    return available
}

// GetRandomOrder returns games in random order
func (gr *GameRegistry) GetRandomOrder() []Game {
    games := gr.GetAllGames()
    gr.rng.Shuffle(len(games), func(i, j int) {
        games[i], games[j] = games[j], games[i]
    })
    return games
}

// GetGameCount returns number of available games
func (gr *GameRegistry) GetGameCount() int {
    return len(gr.GetAllGames())
}