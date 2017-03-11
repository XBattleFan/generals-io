package main

import (
	"fmt"
	"os"
	"os/exec"

	"strings"

	"bytes"

	"github.com/fatih/color"
)

var neutralColor = color.New(color.BgWhite).Add(color.FgBlack)
var playerColor = color.New(color.BgGreen).Add(color.FgBlack)
var playerCityColor = color.New(color.BgGreen).Add(color.FgBlack).Add(color.Underline)
var playerGenColor = color.New(color.BgCyan).Add(color.FgBlack)

var enemyColor = color.New(color.BgRed).Add(color.FgBlack)
var enemyCityColor = color.New(color.BgRed).Add(color.FgBlack).Add(color.Underline)
var enemyGenColor = color.New(color.BgMagenta).Add(color.FgBlack)

func Render(g *Game) {

	size := g.Width * g.Height

	var snapshot bytes.Buffer

	cellDelim := neutralColor.Sprint("|")

	for i := 0; i < size; i++ {
		snapshot.WriteString(cellDelim + fmt.Sprint(terrainToChar(g.Terrain[i], g.Armies[i], ctns(g.Generals, i), ctns(g.rawCities, i), g.PlayerIndex)))
		if (i+1)%g.Width == 0 {
			snapshot.WriteString(neutralColor.Sprintf("\n%s\n", strings.Repeat("-", 7*g.Width)))
		}
	}
	// game score
	snapshot.WriteString(fmt.Sprintf("%s\n%s\nTurn: %d, ", scoreString(g, true), scoreString(g, false), g.turn))
	snapshot.WriteString(fmt.Sprintf("Think time: %v\n", g.ThinkTime))

	// clear terminal
	cmd := exec.Command("clear") //Linux example, its tested
	cmd.Stdout = os.Stdout
	cmd.Run()
	fmt.Println("\n==")

	// vomit the bytes
	fmt.Println(snapshot.String())
}

func scoreString(g *Game, mine bool) string {
	var idx int
	var name string
	if mine {
		idx = g.PlayerIndex
		name = g.PlayerNames[g.PlayerIndex]
	} else {
		idx = g.EnemyIndex
		name = g.PlayerNames[g.EnemyIndex]
	}
	for _, sc := range g.Scores {
		if sc.Index == idx {
			return fmt.Sprintf("%s: [Army: %d | Land: %d] ", name, sc.Armies, sc.Land)
		}
	}
	return ""
}

func ctns(s []int, i int) bool {
	for _, e := range s {
		if e == i {
			return true
		}
	}
	return false
}

func terrainToChar(terrain, army int, gen, city bool, playerIndex int) string {
	switch terrain {
	case fileTile:
		if gen {
			return enemyGenColor.Sprint("  **  ")
		}
		return neutralColor.Sprint("  ~~  ")
	case playerIndex:
		return renderArmy(army, gen, city, false)
	case emptyTile:
		if army > 0 {
			s := fmt.Sprintf("|%d| ", army)
			return neutralColor.Sprintf("%s%s ", strings.Repeat(" ", 5-len(s)), s)
		}
		return neutralColor.Sprint("      ")
	case mountainTile:
		return neutralColor.Sprint("  XX  ")
	case fogObstacleTile:
		if city {
			return neutralColor.Sprint("  CC  ")
		}
		return neutralColor.Sprint("  ??  ")
	default:
		return renderArmy(army, gen, city, true)
	}
}

func renderArmy(army int, gen, city, enemy bool) string {
	s := fmt.Sprintf("%d ", army)
	s = fmt.Sprintf("%s%s ", strings.Repeat(" ", 5-len(s)), s)
	if gen {
		if enemy {
			s = enemyGenColor.Sprint(s)
		} else {
			s = playerGenColor.Sprint(s)
		}
	} else if city {
		if enemy {
			s = enemyCityColor.Sprint(s)
		} else {
			s = playerCityColor.Sprint(s)
		}
	} else {
		if enemy {
			s = enemyColor.Sprint(s)
		} else {
			s = playerColor.Sprint(s)
		}
	}
	return s
}
