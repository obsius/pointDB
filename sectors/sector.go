package sectors

import (
	"encoding/gob"
	"fmt"
	"os"
	"bytes"
	_ "runtime"
	"sync"
)

var mutex = &sync.Mutex{}

type sector struct {

	size    int
	splitAt int

	x     uint32
	y     uint32
	hX    uint32
	hY    uint32
	width uint32

	level uint8

	split bool

	points []*node
	subSectors [2][2]*sector
}

func NewSector(splitAt int) *sector {
	return &sector{
		size:    0,
		splitAt: splitAt,
		split:   false,
		x:       0,
		y:       0,
	}
}

func (this *sector) Append(p *Point) {

	if this.split {

		reducedX := (p.Coord[_x] - this.x) >> (32 - this.level - 1)
		reducedY := (p.Coord[_y] - this.y) >> (32 - this.level - 1)

		this.subSectors[reducedX][reducedY].Append(p)
	} else if this.size >= this.splitAt {

		maxUint32 := uint32(4294967295)
		ranger := uint32(maxUint32 >> uint32(this.level+1))

		for x := uint32(0); x < 2; x++ {
			for y := uint32(0); y < 2; y++ {
				this.subSectors[x][y] = &sector{
					size:    0,
					splitAt: this.splitAt * 2,
					level:   this.level + 1,
					width:   ranger,
					x:       this.x + (x * ranger),
					y:       this.y + (y * ranger),
					hX:      this.x + (x * ranger) + ranger,
					hY:      this.y + (y * ranger) + ranger,
					split:   false,
				}

				newSector := this.subSectors[x][y]

				sX, _ := binarySearch(this.points, this.x+ranger)

				if x == 0 {
					for xI := uint32(0); xI < sX; xI++ {

						src := this.points[xI].nodes
						loc := &this.points[xI].nodes

						sY, _ := binarySearch(src, this.y+ranger)

						var buf []*node
						if y == 0 {
							buf = src[:sY]
							*loc = src[sY:]
						} else if y == 1 {
							buf = src[sY:]
							*loc = src[:sY]
						}

						if buf != nil && len(buf) > 0 {
							newSector.points = append(newSector.points, &node{val: this.points[xI].val, nodes: make([]*node, 0)})
							index := newSector.points[len(newSector.points)-1].nodes
							newSector.points[len(newSector.points)-1].nodes = append(index, buf...)

							newSector.size += len(buf)
						}
					}
				} else {
					for xI := sX; xI < uint32(len(this.points)); xI++ {

						src := this.points[xI].nodes
						loc := &this.points[xI].nodes

						sY, _ := binarySearch(src, this.y+ranger)

						var buf []*node
						if y == 0 {
							buf = src[sY:]
							*loc = src[:sY]
						} else if y == 1 {
							buf = src[:sY]
							*loc = src[sY:]
						}

						if buf != nil && len(buf) > 0 {
							newSector.points = append(newSector.points, &node{val: this.points[xI].val, nodes: make([]*node, 0)})
							index := newSector.points[len(newSector.points)-1].nodes
							newSector.points[len(newSector.points)-1].nodes = append(index, buf...)

							newSector.size += len(buf)
						}
					}
				}
			}
		}

		this.points = nil
		this.split = true

		this.Append(p)
	} else {

		x := p.Coord[_x]
		y := p.Coord[_y]

		xI := this.points

		xIndex, found := binarySearch(xI, x)

		if found {
			yI := xI[xIndex].nodes
			yIndex, found := binarySearch(yI, y)

			if found {
				pI := yI[yIndex].points
				this.points[xIndex].nodes[yIndex].points = append(pI, p)
			} else {
				pI := make([]*Point, 1)
				pI[0] = p
				this.points[xIndex].nodes = insertNode(yI, yIndex, &node{val: y, points: pI})
			}
		} else {
			pI := make([]*Point, 1)
			pI[0] = p
			yI := make([]*node, 1)
			yI[0] = &node{val: y, points: pI}
			this.points = insertNode(xI, xIndex, &node{val: x, nodes: yI})
		}

		this.size++
	}
}

func (this *sector) recurFind(topLeft Coord, bottomRight Coord) []*sector {

	var retSectors []*sector

	if !this.split {
		retSectors = append(retSectors, this)
	} else {
		for x := uint32(0); x < 2; x++ {
			for y := uint32(0); y < 2; y++ {
				subSector := this.subSectors[x][y]
				if overlap(subSector.x, subSector.hX, topLeft[_x], bottomRight[_x]) && overlap(subSector.y, subSector.hX, topLeft[_y], bottomRight[_y]) {
					retSectors = append(retSectors, subSector.recurFind(topLeft, bottomRight)...)
				}
			}
		}
	}

	return retSectors
}

func (this *sector) FindInSector(query Query) TileMatrix {

	var wg sync.WaitGroup

	// find all master sectors
	var sectors = this.recurFind(query.TopLeft, query.BottomRight)

	fmt.Println(sectors)

	out := make(chan FindSpec, 4)
	in := make(chan TileMatrix, len(sectors))
	tileMatrix := make(TileMatrix)

	for b := 0; b < 4; b++ {
		wg.Add(1)
		go worker(out, in, &wg)
	}

	for i := 0; i < len(sectors); i++ {
		out <- FindSpec{
			topLeft:     query.TopLeft,
			bottomRight: query.BottomRight,
			sector:      sectors[i],
			Level:       query.Level,
		}
	}

	close(out)

	wg.Wait()
	close(in)

	for tM := range in {
		mergeTileMatrices(tileMatrix, tM)
	}

	return tileMatrix
}

type TileMatrix map[uint32]map[uint32]*Tile

type FindSpec struct {
	topLeft     Coord
	bottomRight Coord

	Level  uint8
	sector *sector
}

func (this *sector) PrintSector() {
	if this.split {
		for x := range this.subSectors {
			for y := range this.subSectors[x] {
				this.subSectors[x][y].PrintSector()
			}
		}
	} else {
		fmt.Printf("{x=%v, y=%v, s=%v} %v\n", this.x, this.y, "unk", this.size)
	}
}

func populateTiles(level uint8, points []*Point) TileMatrix {

	tiles := make(TileMatrix)

	for i := range points {

		x := points[i].Coord[_x] >> level
		y := points[i].Coord[_y] >> level

		if _, ok := tiles[x]; !ok {
			tiles[x] = make(map[uint32]*Tile)
		}

		if _, ok := tiles[x][y]; !ok {
			tiles[x][y] = &Tile{
				X:     x,
				Y:     y,
				Count: 0,
			}
		}

		tiles[x][y].Count++
	}

	return tiles
}

func mergeTileMatrices(dest TileMatrix, new TileMatrix) {
	for x := range new {
		if _, ok := dest[x]; ok {
			for y := range new[x] {
				if _, ok := dest[x][y]; ok {
					dest[x][y].Count += new[x][y].Count
				} else {
					dest[x][y] = new[x][y]
				}
			}
		} else {
			dest[x] = new[x]
		}
	}
}

func worker(in chan FindSpec, out chan TileMatrix, wg *sync.WaitGroup) {

	defer wg.Done()

	for spec := range in {
		tileMatrix := make(TileMatrix)

		xI := spec.sector.points
		sector := spec.sector

		if sector.x >= spec.topLeft[_x] && sector.hX <= spec.bottomRight[_x] &&
			sector.y >= spec.topLeft[_y] && sector.hY <= spec.bottomRight[_y] {
			for x := 0; x < len(sector.points); x++ {
				for y := 0; y < len(sector.points[x].nodes); y++ {
					yI := sector.points[x].nodes[y]
					mergeTileMatrices(tileMatrix, populateTiles(spec.Level, yI.points))
				}
			}
		} else {
			lowX, _ := binarySearch(xI, spec.topLeft[_x])
			highX, _ := binarySearch(xI, spec.bottomRight[_x])

			for x := lowX; x < highX; x++ {
				yI := xI[x].nodes

				low, _ := binarySearch(yI, spec.topLeft[_y])
				high, _ := binarySearch(yI, spec.bottomRight[_y])

				for y := low; y < high; y++ {
					mergeTileMatrices(tileMatrix, populateTiles(spec.Level, yI[y].points))
				}
			}
		}

		out <- tileMatrix
	}
}

func (this *sector) GobEncode() ([]byte, error) {

    w := new(bytes.Buffer)

    encoder := gob.NewEncoder(w)
    err := encoder.Encode(this.points)
    if err != nil { return nil, err }
	if this.split {
    	err = encoder.Encode(this.subSectors)
	}
    if err != nil { return nil, err }

    return w.Bytes(), nil
}

func (this *sector) GobDecode(buf []byte) error {

    r := bytes.NewBuffer(buf)
    decoder := gob.NewDecoder(r)

    err := decoder.Decode(this.points)
    if err != nil { return err }
	err = decoder.Decode(this.subSectors)
    if err != nil { return err }

    return nil
}

func (this *sector) WriteToFile(filename string) error {

	fmt.Printf("%v", len(this.points))

	file, err := os.Create(filename)
	if err == nil {
		encoder := gob.NewEncoder(file)
		encoder.Encode(this)
	}
	file.Close()
	return err
}
