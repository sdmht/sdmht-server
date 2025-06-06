package game

import (
	"log"
	"sync"

	"github.com/emirpasic/gods/lists/arraylist"
	"github.com/olebedev/emitter"
)

type Player struct {
	Size    int32
	Version string
}

type PlayerMap map[string]Player

type Game struct {
	online_uids    arraylist.List
	Player         PlayerMap
	MatchingPlayer PlayerMap
	Pmu            sync.RWMutex
	Event          emitter.Emitter
}

func NewGame() *Game {
	return &Game{
		Player:         make(PlayerMap),
		MatchingPlayer: make(PlayerMap),
	}
}

func (g *Game) OnlineCount() int32 {
	return int32(g.online_uids.Size())
}

func (g *Game) OnlineChanged() {
	v := g.OnlineCount()
	log.Print(v, " 在线")
	go g.Event.Emit("online_changed", v)
}

func (g *Game) Join(uid string) {
	log.Print(uid, " 来访")
	g.online_uids.Add(uid)
	g.OnlineChanged()
}

func (g *Game) Leave(uid string) {
	log.Print(uid, " 离开")
	g.online_uids.Remove(g.online_uids.IndexOf(uid))
	g.OnlineChanged()
	go g.Event.Emit("leave" + uid)
}

func (g *Game) IsOnline(uid string) bool {
	return g.online_uids.Contains(uid)
}
