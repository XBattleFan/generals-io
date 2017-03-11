package main

import (
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"testing"
	"time"
)

func TestSpeed(t *testing.T) {
	start := time.Now()

	var wg sync.WaitGroup
	wg.Add(500)
	var final []int

	for i := 0; i < 500; i++ {
		go func() {
			defer wg.Done()
			var nums []int
			for i := 0; i < 10000000/500; i++ {
				nums = append(nums, i)
			}
			final = combine(nums, final)
		}()
	}
	wg.Wait()
	fmt.Println(len(final))
	sort.Sort(sort.IntSlice(final))
	sort.Reverse(sort.IntSlice(final))
	sort.Sort(sort.IntSlice(final))
	sort.Reverse(sort.IntSlice(final))
	sort.Sort(sort.IntSlice(final))
	sort.Reverse(sort.IntSlice(final))
	sort.Sort(sort.IntSlice(final))
	sort.Reverse(sort.IntSlice(final))
	fmt.Println(final[rand.Intn(len(final)-1)])

	elapsed := time.Since(start)
	fmt.Printf("This shit took %v\n", elapsed)
}

var cm = &sync.Mutex{}

func combine(src, dst []int) []int {
	cm.Lock()
	defer cm.Unlock()
	dst = append(dst, src...)
	return dst
}

type Person struct {
	name string
	Age  int
}

func TestRounding(t *testing.T) {
	people := []Person{
		{"Bob", 31},
		{"John", 42},
		{"Michael", 17},
		{"Jenny", 26},
	}

	fmt.Println(people)
	sort.Slice(people, func(i, j int) bool { return people[i].Age < people[j].Age })
	fmt.Println(people)
}

func TestPathfinding(t *testing.T) {
	pI := 0
	armies := []int{
		0, 0, 0,
		5, 1, 0,
		0, 0, 0}
	terrain := []int{
		mountainTile, mountainTile, mountainTile,
		pI, pI, emptyTile,
		mountainTile, mountainTile, mountainTile}

	game := Game{Armies: armies, Terrain: terrain, Width: 3, Height: 3}
	brain := Brain{g: &game}

	brain.Think()
}

func TestSet(t *testing.T) {
	set := map[int]struct{}{}

	set[6] = struct{}{}
	if _, ok := set[3]; ok {
		t.Error(ok)
	}

	if _, ok := set[5]; !ok {
		t.Error(ok)
	}

}

func TestThirdLast(t *testing.T) {
	s := []int{1, 2, 3, 4}
	fmt.Println(s[len(s)-3:])

}
