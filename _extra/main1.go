package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"
	"sync"
)

var mutex = &sync.Mutex{}

const _x uint8 = 0
const _y uint8 = 1

const numTrials = 10

type Point struct {
	Coord [2]uint32
}

type Node struct {
	val uint32
	data interface{}
}

type PointDB struct {
	points []*Node
}

type dd struct {
	lowerX uint32
	upperX uint32
	lowerY uint32
	upperY uint32
	i int
}

type MasterSector struct {
	x uint32
	y uint32

	power uint8
}

type Sector struct {
	x uint32
	y uint32

	Count uint32
}

func binarySearch(nodes []*Node, val uint32) (uint32, bool) {

	length := len(nodes)

	count := int(math.Ceil(math.Log2(float64(length))))
	cc := 0

	found := false

	var i int = len(nodes) / 2
	var c float64;
	for c = float64(float64(length) / 4); cc <= count; c = c / 2 {

		if (i < 0) { i = 0 }
		if (i > length - 1) { i = length - 1 }

		if val > nodes[i].val {
			i = i + int(math.Ceil(c))
		} else if val < nodes[i].val && c >= .5  {
			i = i - int(math.Ceil(c))
		} else if (val == nodes[i].val) {
			found = true
			break
		} else {
			break
		}
		cc++
	}

	return uint32(i), found
}

func insertNode(arr []*Node, i uint32, val *Node) []*Node {

	arr = append(arr, &Node{})
	copy(arr[i+1:], arr[i:])
	arr[i] = val

	return arr
}

func (pDB *PointDB) append(in chan *Point, i uint32, wg *sync.WaitGroup) {

	defer wg.Done()

	for p := range in {

		x := p.Coord[_x]
		y := p.Coord[_y]

		node := pDB.points[i]

		if len(node.data.([]*Node)) >= int(node.val) {

			mutex.Lock()

			node.val *= 2
			
			newNode := Node{ val: node.val }

			src := node.data.([]*Node)
			destI := len(pDB.points)

			fmt.Printf("%v->%v, %v\n", i, destI, len(src))

			pDB.points[i].data = src[len(src) / 2 : len(src)]

			newNode.data = src[0 : len(src) / 2]

			fmt.Printf("%v    %v\n", len(pDB.points[i].data.([]*Node)), len(newNode.data.([]*Node)))

			pDB.points = append(pDB.points, &newNode)

			wg.Add(1)
			go pDB.append(in, uint32(destI), wg)

			mutex.Unlock()
		}

		xI := pDB.points[i].data.([]*Node)

		xIndex, found := binarySearch(xI, x)

		if found {
			yI := xI[xIndex].data.([]*Node)
			yIndex, found := binarySearch(yI, y)

			if (found) {
				pI := yI[yIndex].data.([]*Point)
				pDB.points[i].data.([]*Node)[xIndex].data.([]*Node)[yIndex].data = append(pI, p)
			} else {
				pI := make([]*Point, 1)
				pI[0] = p
				pDB.points[i].data.([]*Node)[xIndex].data = insertNode(yI, yIndex, &Node{ val: y, data: pI })
			}
		} else {
			pI := make([]*Point, 1)
			pI[0] = p
			yI := make([]*Node, 1)
			yI[0] = &Node{ val: y, data: pI }
			pDB.points[i].data = insertNode(xI, xIndex, &Node{ val: x, data: yI })
		}
	}
}

func (pDB *PointDB) appendMaster() {
	var wg sync.WaitGroup

	out := make(chan *Point, 100)

	wg.Add(1)
	go pDB.append(out, 0, &wg)

	for c := 0; c < 1000000; c++ {
		out <- &Point{ [2]uint32{ rand.Uint32(), rand.Uint32() } }
	}

	close(out)

	wg.Wait()
}

func (pDB *PointDB) worker(in chan dd, out chan []*Point, wg *sync.WaitGroup) {

	defer wg.Done()

	for data := range in {
		var points []*Point

		xI := pDB.points[data.i].data.([]*Node)
		
		lowX, _ := binarySearch(xI, data.lowerX)
		highX, _ := binarySearch(xI, data.upperX)

		for x := lowX; x < highX; x++ {
			yI := xI[x].data.([]*Node)

			low, _ := binarySearch(yI, data.lowerY)
			high, _ := binarySearch(yI, data.upperY)

			for y := low; y < high; y++ {
				points = append(points, yI[y].data.([]*Point)...)
			}
		}

		var countt = 0
		for xx := range pDB.points[data.i].data.([]*Node) {
			xxI := pDB.points[data.i].data.([]*Node)
			for yy := range xxI[xx].data.([]*Node) {
				yyI := xI[xx].data.([]*Node)
				countt += len(yyI[yy].data.([]*Point))
			}
		}
		fmt.Println(countt)
		fmt.Println(len(points))

		out <- points
	}
}

func (pDB *PointDB) findInSector(lowerX uint32, upperX uint32, lowerY uint32, upperY uint32) []*Point {

	var wg sync.WaitGroup

	out := make(chan dd, 4)
	in := make(chan []*Point, len(pDB.points))
	var points []*Point

	for b := 0; b < 4; b++ {
		wg.Add(1)
		go pDB.worker(out, in, &wg)
	}

	fmt.Println(len(pDB.points))

	for i := 0; i < len(pDB.points); i++ {
		out <- dd {
			lowerX: lowerX,
			lowerY: lowerY,
			upperX: upperX,
			upperY: upperY,
			i: i,
		}
	}

	close(out)

	wg.Wait()
	close(in)

	for ps := range in {
		points = append(points, ps...)
	}

	zoom := uint32(31)
	sectors := make(map[uint32]map[uint32]*Sector)
	for i := range points {

		x := points[i].Coord[_x] >> zoom
		y := points[i].Coord[_y] >> zoom

		if _, ok :=sectors[x]; !ok {
			sectors[x] = make(map[uint32]*Sector)
		}

		if _, ok := sectors[x][y]; !ok {
			sectors[x][y] = &Sector{
				x: x,
				y: y,
				Count: 0,
			}
		}

		sectors[x][y].Count++
	}

	for x := range sectors {
		for y := range sectors[x] {
			fmt.Printf("%v", *sectors[x][y])
		}
		fmt.Println()
	}

	return points
}

func main() {

	points := PointDB{
		points: make([]*Node, 1),
	}

	points.points[0] = &Node{ val: 16384, data: make([]*Node, 1) }
	points.points[0].data = points.points[0].data.([]*Node)[:0]

	// inserts
	start := time.Now()
	points.appendMaster()
	fmt.Printf("Insertion time: %s\n", time.Since(start))

/*

	fmt.Println(len(points.points))
	for i := 0; i < len(points.points); i++ {
		fmt.Println(len(points.points[i].data.([]*Node)))
	}

	dataI := points.points[0].data.([]*Node)
	fmt.Println(len(dataI))
	for x := 0; x < len(dataI); x++ {
		dataX := dataI[x].data.([]*Node)
		for y := 0; y < len(dataX); y++ {
			dataY := dataX[y].data.([]*Point)
			for i := 0; i < len(dataY); i++ {
				fmt.Print(dataY[i])
			}
		}
		fmt.Println()
	}
*/

	uintSize := uint32(4294967295)

	// multi-thread
	start = time.Now()
	var buf []*Point
	for i := 0; i < 1; i++ {
		buf = points.findInSector(0, uintSize / 16, 0, uintSize / 16)
	}
	fmt.Printf("found: %v\n", len(buf))
	fmt.Printf("Random Read time: %s\n", time.Since(start))
}
