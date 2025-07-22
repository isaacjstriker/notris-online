package games

import (
	"math/rand"
	"time"

	"github.com/isaacjstriker/devware/games/tetris"
	"github.com/isaacjstriker/devware/games/typing"
	"github.com/isaacjstriker/devware/internal/types"
)

// GameRegistry manages all available games
type GameRegistry struct {
	games []types.Game
	rng   *rand.Rand
}

// NewGameRegistry creates a new game registry
func NewGameRegistry() *GameRegistry {
	return &GameRegistry{
		games: make([]types.Game, 0),
		rng:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// RegisterGame adds a game to the registry
func (gr *GameRegistry) RegisterGame(game types.Game) {
	gr.games = append(gr.games, game)
}

// GetAllGames returns all registered games
func (gr *GameRegistry) GetAllGames() []types.Game {
	available := make([]types.Game, 0)
	for _, game := range gr.games {
		if game.IsAvailable() {
			available = append(available, game)
		}
	}
	return available
}

// GetRandomOrder returns games in random order
func (gr *GameRegistry) GetRandomOrder() []types.Game {
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

// Registry of available games
var Games = map[string]func() types.Game{
	"typing": func() types.Game { return typing.NewTypingGame() },
	"tetris": func() types.Game { return tetris.NewTetris() },
}

// GetGameList returns a list of available games
func GetGameList() []string {
	var games []string
	for name := range Games {
		games = append(games, name)
	}
	return games
}
