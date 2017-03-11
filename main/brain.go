package main

import (
	"errors"
	"fmt"
	"time"

	"sort"

	"math"

	"github.com/michalbric/generals-io/opti"
)

// Brain does the thinking
type Brain struct {
	c *Client
	g *Game
}

// Think decides what the next move is
func (b *Brain) Think() {
	start := time.Now()

	if b.g.turn < 20 {
		return
	}

	f, t := b.pickActors()
	allPaths, err := b.paths(f, t)
	if err != nil {
		return
	}

	chosenPath, halfAttack := b.pickPath(allPaths)
	b.c.Attack(chosenPath.start, chosenPath.firstStep, halfAttack)
	b.g.ThinkTime = time.Since(start)
}

func (b *Brain) pickActors() ([]int, []int) {
	friendly, enemy, neutral := b.classifiedTiles()

	if len(friendly) > 10 {
		sort.Slice(friendly, func(i, j int) bool {
			return b.armyAt(friendly[i]) < b.armyAt(friendly[j])
		})
		friendly = friendly[len(friendly)-10:]
	}
	targets := append(neutral, enemy...)

	return friendly, targets
}

func (b *Brain) pickPath(paths []path) (path, bool) {
	sort.Slice(paths, func(i, j int) bool {
		return b.pathValue(paths[i]) < b.pathValue(paths[j])
	})
	chosenPath := paths[len(paths)-1]
	return chosenPath, false
}

func (b *Brain) pathValue(p path) int {
	var bonus int
	if genIndex := b.g.Generals[b.g.EnemyIndex]; genIndex > 0 {
		if genIndex == p.finalTarget {
			bonus += 99999999
		} else if genDst := b.distance(genIndex, p.finalTarget); genDst < 5 {
			bonus += 800 - genDst*100
		}
	}
	/*
		if b.isEnemy(p.finalTarget) && b.distance(b.g.Generals[b.g.PlayerIndex], p.finalTarget) < 6 {
			bonus += 350
		} else {
			bonus += b.distance(b.g.Generals[b.g.PlayerIndex], p.firstStep) * 8
		}
		if b.isEnemy(p.firstStep) {
			bonus += 75
			if b.isCity(p.firstStep) {
				bonus += 100
			}
		}

		if b.isEnemy(p.finalTarget) {
			bonus += b.armyAt(p.finalTarget) * 2
		}

		return p.finalArmy*3 + b.armyAt(p.start) + p.armyKilled*10 - p.dist*15 + bonus // reduced dist from *10
	*/
	return int((float64(p.finalArmy+p.armyKilled*2)/float64(p.dist))*1000) + bonus
}

// path contains the first step of the path and metainfo about the whole path
type path struct {
	start           int
	firstStep       int
	finalTarget     int
	dist            int
	neutralCaptures int
	enemyCaptures   int
	armyKilled      int
	finalArmy       int
}

func (p path) String() string {
	return fmt.Sprintf("{st:%d, fs:%d, ft:%d, dist:%d, nc:%d, ec:%d, ak:%d, fa:%d}",
		p.start, p.firstStep, p.finalTarget, p.dist, p.neutralCaptures, p.enemyCaptures, p.armyKilled, p.finalArmy)
}

// paths currently find the shortest paths between friendlies and targets
// while the bfs walks from targets to friendlies, the path struct uses intuitive
// from->to, where from will always be a walkable friendly
func (b *Brain) paths(friendly, targets []int) ([]path, error) {
	var allPaths []path
	unvisitedTiles := opti.NewQueue()

	for _, f := range friendly {
		for _, t := range targets {
			shortestDist := math.MaxInt32
			visitedTiles := make(map[int]struct{})
			unvisitedTiles.Add(path{f, -1, f, 0, 0, 0, 0, b.signedArmyAt(f)})

			for unvisitedTiles.Length() > 0 {
				curr, _ := unvisitedTiles.Remove().(path)
				if curr.finalArmy-curr.dist < 1 {
					continue
				}
				if curr.finalTarget == t {
					if curr.finalArmy-curr.dist > 0 && curr.dist <= shortestDist {
						allPaths = append(allPaths, curr)
						shortestDist = curr.dist
					}
					//continue // no more unwrapping
				}
				// add unvisited neighbors
				for _, next := range b.neighbors(curr.finalTarget) {
					if _, v := visitedTiles[next]; v || !b.visibleUnit(next) {
						continue // skip seen and obstacles
					}

					// assign first step if necessary
					nextFirstStep := curr.firstStep
					if nextFirstStep < 0 {
						nextFirstStep = next
					}

					visitedTiles[next] = struct{}{} // mark next as visited

					toSignedArmy := b.signedArmyAt(next)
					nextDist := curr.dist + 1
					nextNeutralCaptures := curr.neutralCaptures
					nextEnemyCaptures := curr.enemyCaptures
					nextArmyKilled := curr.armyKilled
					nextFinalArmy := curr.finalArmy + toSignedArmy

					if !b.isMine(next) {
						if b.isEnemy(next) {
							nextEnemyCaptures += 1
						} else { // neutral tile
							nextNeutralCaptures += 1
						}
						nextArmyKilled -= toSignedArmy
					}
					unvisitedTiles.Add(path{
						f, nextFirstStep, next, nextDist, nextNeutralCaptures,
						nextEnemyCaptures, nextArmyKilled, nextFinalArmy})

				}

			}
		}
	}

	if len(allPaths) == 0 {
		return allPaths, errors.New("No path found")
	}
	return allPaths, nil
}

func (b *Brain) armyAt(i int) int {
	return b.g.Armies[i]
}

func (b *Brain) signedArmyAt(i int) int {
	if b.isMine(i) {
		return b.armyAt(i)
	}
	return b.armyAt(i) * -1
}

func (b *Brain) isMine(i int) bool {
	return b.g.Terrain[i] == b.g.PlayerIndex
}

func (b *Brain) isEnemy(i int) bool {
	return b.g.Terrain[i] == b.g.EnemyIndex
}

func (b *Brain) isNeutral(i int) bool {
	return b.g.Terrain[i] == emptyTile
}

func (b *Brain) isCity(i int) bool {
	_, ok := b.g.CityMap[i]
	return ok
}

// shamelessly stolen from @vendan
func (b *Brain) distance(from, to int) int {
	x1, y1 := from%b.g.Width, from/b.g.Width
	x2, y2 := to%b.g.Width, to/b.g.Width
	dx := x1 - x2
	dy := y1 - y2
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	return dx + dy
}

func (b *Brain) neighbors(pos int) []int {
	ns := make([]int, 0, 4)
	if pos >= b.g.Width {
		ns = append(ns, pos-b.g.Width)
	}
	if pos < b.g.Width*(b.g.Height-1) {
		ns = append(ns, pos+b.g.Width)
	}
	if pos%b.g.Width > 0 {
		ns = append(ns, pos-1)
	}
	if pos%b.g.Width < b.g.Width-1 {
		ns = append(ns, pos+1)
	}
	return ns
}

func (b *Brain) visibleUnit(i int) bool {
	tileType := b.g.Terrain[i]
	return tileType != fileTile && tileType != fogObstacleTile && tileType != mountainTile
}

// classifiedTiles distributes the whole visible terrain into groups of neutrals, friendlies and enemies
func (b *Brain) classifiedTiles() (friendly, enemy, neutral []int) {
	for idx := range b.g.Terrain {
		if b.visibleUnit(idx) {
			if b.isMine(idx) { // we do not care about non-movable tiles
				friendly = append(friendly, idx)
			} else if b.isNeutral(idx) {
				neutral = append(neutral, idx)
			} else {
				enemy = append(enemy, idx)
			}
		}
	}
	return
}
