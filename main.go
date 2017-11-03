package main

import (
	"fmt"
	"math/rand"
	"time"

	"./sectors"
)

const _x uint8 = 0
const _y uint8 = 1

func main() {

	root := sectors.NewSector(4096)

	// inserts
	start := time.Now()
	
	for i := uint32(0); i < 200000; i++ {
		root.Append(&sectors.Point{
			ID: i,
			Coord: [2]uint32{ rand.Uint32(), rand.Uint32() },
		})
	}

	fmt.Printf("Insertion time: %s\n", time.Since(start))

	root.PrintSector()

	uintSize := uint32(4294967295)

	// multi-thread
	start = time.Now()
	var buf sectors.TileMatrix
	for i := 0; i < 1; i++ {
		buf = root.FindInSector(sectors.Query{
			TopLeft: sectors.Coord{0, 0},
			BottomRight: sectors.Coord{uintSize / 1, uintSize / 1},
			Level: 31,
		})
	}

	//json, _ := json.Marshal(buf)

	for x := range buf {
		for y := range buf[x] {
			fmt.Printf("{%v %v}: %v\n", x, y, (*buf[x][y]).Count)
		}
	}
	//fmt.Printf("found: %v\n", json)
	fmt.Printf("found: %v\n", len(buf))
	fmt.Printf("Random Read time: %s\n", time.Since(start))

	root.WriteToFile("test.bin")
}
