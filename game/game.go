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
	online_uids     arraylist.List
	Player          PlayerMap
	MatchingPlayer  PlayerMap
	Pmu             sync.RWMutex
	Event           emitter.Emitter
	cachedResources map[string]map[string]struct{}
	crmu            sync.RWMutex
}

func NewGame() *Game {
	g := &Game{
		Player:          make(PlayerMap),
		MatchingPlayer:  make(PlayerMap),
		cachedResources: make(map[string]map[string]struct{}),
	}
	g.Event.Use("*", emitter.Void)
	return g
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
	g.RemoveCachedResources(uid)
	go g.Event.Emit("leave:" + uid)
}

func (g *Game) IsOnline(uid string) bool {
	return g.online_uids.Contains(uid)
}

func (g *Game) PublishCachedResources(uid string, paths []string) {
	g.crmu.Lock()
	defer g.crmu.Unlock()
	for _, path := range paths {
		if _, ok := g.cachedResources[path]; !ok {
			g.cachedResources[path] = make(map[string]struct{})
		}
		g.cachedResources[path][uid] = struct{}{}
	}
}

func (g *Game) RemoveCachedResources(uid string) {
	g.crmu.Lock()
	defer g.crmu.Unlock()
	for path, uids := range g.cachedResources {
		delete(uids, uid)
		if len(uids) == 0 {
			delete(g.cachedResources, path)
		}
	}
}

func (g *Game) GetCachedResourcePeers(uid string, path string) []string {
	g.crmu.RLock()
	defer g.crmu.RUnlock()
	uids, ok := g.cachedResources[path]
	if !ok {
		return []string{}
	}
	peers := make([]string, 0, len(uids)-1)
	for u := range uids {
		if u != uid {
			peers = append(peers, u)
		}
	}
	return peers
}
