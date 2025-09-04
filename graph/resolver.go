package graph

import (
	"sdmht-server/game"
)

type Resolver struct {
	game game.Game
}

func NewResolver() *Resolver {
	return &Resolver{
		game: *game.NewGame(),
	}
}
