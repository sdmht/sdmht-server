package game

import (
	"log"
	"sync"

	"github.com/btvoidx/mint"
	"github.com/emirpasic/gods/lists/arraylist"
)

type Player struct {
	Size        int32
	Version     string
	DataChannel chan any
}

type PlayerMap map[string]Player

type Game struct {
	online_uids    arraylist.List
	online_changed mint.Emitter
	leave          mint.Emitter
	Player         PlayerMap
	Pmu            sync.RWMutex
	MatchingPlayer PlayerMap
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
	mint.Emit(&g.online_changed, v)
}

func (g *Game) OnOnlineChanged(fn func(int32)) (off func() <-chan struct{}) {
	return mint.On(&g.online_changed, fn)
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
	mint.Emit(&g.leave, uid)
}

func (g *Game) IsOnline(uid string) bool {
	return g.online_uids.Contains(uid)
}

func (g *Game) OnLeave(uid string, ch chan any) (off func() <-chan struct{}) {
	go func() {
		if !g.IsOnline(uid) {
			ch <- nil
		}
	}()
	return mint.On(&g.leave, func(luid string) {
		if uid == luid {
			ch <- nil
		}
	})
}
