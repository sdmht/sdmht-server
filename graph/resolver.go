package graph

import (
	"sdmht-server/db"
	"sdmht-server/game"

	"gorm.io/gorm"
)

type Resolver struct {
	game game.Game
	db   *gorm.DB
}

func NewResolver() *Resolver {
	return &Resolver{
		game: *game.NewGame(),
		db:   db.DB,
	}
}
