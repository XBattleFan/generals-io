package main

import (
	"log"

	"os"
	"sync"
)

func main() {
	// defer profile.Start(profile.MemProfile).Stop()
	var wg sync.WaitGroup
	wg.Add(1)
	c, err := NewClient()
	if err != nil {
		log.Fatal(err)
	}
	go c.Connect()

	if len(os.Args) == 2 {
		id := os.Args[1]
		c.JoinPrivate(id)
		c.ForceStart(id)
	} else {
		c.Join1v1()
	}

	wg.Wait()
}

//TODO notes ------------
/*

@vendan wrote:
first round of "pathfinding" is basically a modified floodfill from every cell
the value's it's flooding are basically a distance modified "weighting"
so that it weighs tiles near enemies higher and such
then, it does a collection path find from every non-owned cell, identifying a "collection path" that'll collect enough cells to capture that tile
then it combines the 2, weighted on how long it'll take to take the tile, and how desired the tile is
highest total score is identified, then the first step of it's collection path is performed
then it throws it all away and does it again the next update

//TODO first -> clean the old methods, comment rules VERY BASIC meaning - e.g. land grab less in later turns -> then refactor
//TODO decrease land grab value lategame
//TODO rework the devalueFreq -> or give very strong pull towards attacking, maybe score average distance to enemy tiles
//TODO add value to moving towards enemy
//TODO: check if general discovered (bool that is set to true if unit next to it) -> protect gen after

*/
