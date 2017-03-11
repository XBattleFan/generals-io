package main

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	emptyTile       int = -1
	mountainTile    int = -2
	fileTile        int = -3
	fogObstacleTile int = -4
)

type GameUpdate struct {
	AttackIndex int         `json:"attackIndex"`
	Turn        int         `json:"turn"`
	MapDiff     []int       `json:"map_diff"`
	CitiesDiff  []int       `json:"cities_diff"`
	Generals    []int       `json:"generals"`
	Scores      []GameScore `json:"scores"`
}

type GameStart struct {
	PlayerIndex int      `json:"playerIndex"`
	ReplayID    string   `json:"replay_id"`
	Usernames   []string `json:"usernames"`
}

type GameScore struct {
	Armies int  `json:"total"`
	Land   int  `json:"tiles"`
	Index  int  `json:"i"`
	Dead   bool `json:"dead"`
}

// Game holds the current state as well as pointers to brain and client
type Game struct {
	c         *Client
	b         *Brain
	ThinkTime time.Duration

	PlayerNames []string
	PlayerIndex int
	EnemyIndex  int

	ReplayID string
	Scores   []GameScore
	gameID   string
	turn     int

	GameMap []int
	Height  int
	Width   int

	Armies      []int
	Terrain     []int
	Generals    []int
	rawCities   []int
	CityMap     map[int]struct{}
	neighborMap map[int][]int
}

func NewGame(c *Client, gameID string) *Game {
	g := &Game{c: c, gameID: gameID}
	g.b = &Brain{c: c, g: g}

	// regiser event handlers
	g.c.Handlers[UpdateEvent] = func(data json.RawMessage) {
		g.Update(data)
		go func() {
			g.b.Think()
		}()
		go func() {
			Render(g)
		}()
	}
	g.c.Handlers[StartEvent] = func(data json.RawMessage) {
		g.Start(data)
	}
	g.c.Handlers[LostEvent] = func(data json.RawMessage) {
		g.GameEnded(data, false)
	}
	g.c.Handlers[WonEvent] = func(data json.RawMessage) {
		g.GameEnded(data, true)
	}
	g.c.Handlers[QueueEvent] = func(data json.RawMessage) {
		fmt.Println("Queue updated")
	}
	g.c.Handlers[PreStartEvent] = func(data json.RawMessage) {
		fmt.Println("Game Starting...")
	}
	return g
}

// Update is called on every game gameUpdate
func (g *Game) Update(updateJSON json.RawMessage) {
	var gameUpdate GameUpdate
	decode := []interface{}{nil, &gameUpdate}
	json.Unmarshal(updateJSON, &decode)

	g.GameMap = g.patch(g.GameMap, gameUpdate.MapDiff)
	g.updateCities(gameUpdate.CitiesDiff)
	g.updateGenerals(gameUpdate.Generals)

	g.Scores = gameUpdate.Scores
	g.turn = gameUpdate.Turn

	g.Width = g.GameMap[0]
	g.Height = g.GameMap[1]
	size := g.Height * g.Width

	g.Armies = g.GameMap[2 : size+2]
	g.Terrain = g.GameMap[size+2:]

}

// Start is called on StartEvent
func (g *Game) Start(startJSON json.RawMessage) {
	var gameStart GameStart
	decode := []interface{}{nil, &gameStart}
	json.Unmarshal(startJSON, &decode)

	g.PlayerIndex = gameStart.PlayerIndex
	g.ReplayID = gameStart.ReplayID
	g.PlayerNames = gameStart.Usernames
	g.CityMap = make(map[int]struct{})
	g.neighborMap = make(map[int][]int)
	g.Generals = []int{-1, -1}

	// this is ugly but for 1v1 it works
	if g.PlayerIndex == 0 {
		g.EnemyIndex = 1
	} else {
		g.EnemyIndex = 0
	}
}

// GameEnded is called on LostEvent and WonEvent
func (g *Game) GameEnded(endJSON json.RawMessage, won bool) {
	var result string
	if won {
		result = "won"
	} else {
		result = "lost"
	}

	fmt.Println("==================================================")
	fmt.Printf("Game %s against %s\n", result, g.PlayerNames[g.EnemyIndex])
	fmt.Printf("Replay: http://bot.generals.io/replays/%s\n", g.ReplayID)
	fmt.Println("==================================================")

	g.c.sendMsg(leave)
	g.c.Stop()
}

func (g *Game) updateCities(citiesDiff []int) {
	g.rawCities = g.patch(g.rawCities, citiesDiff)
	for _, e := range g.rawCities {
		g.CityMap[e] = struct{}{}
	}
}

// persists generals -> no updaeting after index is known
func (g *Game) updateGenerals(generals []int) {
	for i, e := range generals {
		if e != -1 {
			g.Generals[i] = e
		}
	}
}

func (g *Game) patch(old, diff []int) []int {
	out := make([]int, 0, 2+g.Width*g.Height)
	i := 0

	for i < len(diff) {
		if diff[i] > 0 {
			out = append(out, old[len(out):len(out)+diff[i]]...)
		}
		i++
		if i < len(diff) && diff[i] > 0 {
			out = append(out, diff[i+1:i+diff[i]+1]...)
			i += diff[i]
		}
		i++
	}
	return out
}

func (g *Game) neighbors(idx int) []int {
	ns, ok := g.neighborMap[idx]
	if !ok {
		ns = make([]int, 0, 4)
		if idx >= g.Width {
			ns = append(ns, idx-g.Width)
		}
		if idx < g.Width*(g.Height-1) {
			ns = append(ns, idx+g.Width)
		}
		if idx%g.Width > 0 {
			ns = append(ns, idx-1)
		}
		if idx%g.Width < g.Width-1 {
			ns = append(ns, idx+1)
		}
		g.neighborMap[idx] = ns
	}

	return ns
}
