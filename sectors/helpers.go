package sectors

import (
	"math"
)

type Coord [2]uint32
type Level uint8

const _x uint8 = 0
const _y uint8 = 1

func binarySearch(nodes []*node, val uint32) (uint32, bool) {

	length := len(nodes)

	if length == 0 {
		return 0, false
	}

	count := int(math.Ceil(math.Log2(float64(length))))
	cc := 0

	found := false

	var i int = len(nodes) / 2
	var c float64
	for c = float64(float64(length) / 4); cc <= count; c = c / 2 {

		if i < 0 {
			i = 0
		}
		if i > length-1 {
			i = length - 1
		}

		if val > nodes[i].val {
			i = i + int(math.Ceil(c))
		} else if val < nodes[i].val && c >= .5 {
			i = i - int(math.Ceil(c))
		} else if val == nodes[i].val {
			found = true
			break
		} else {
			break
		}
		cc++
	}

	return uint32(i), found
}

func overlap(aLow uint32, aHigh uint32, bLow uint32, bHigh uint32) bool {
	return (aLow >= bLow && aLow <= bHigh) || (bLow >= aLow && bLow <= aHigh) || (aHigh >= bLow && aHigh <= bHigh) || (bHigh >= aLow && bHigh <= aHigh)
}
