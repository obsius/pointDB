package sectors

import (
	"encoding/gob"
	_"encoding/binary"
	"bytes"
)

type node struct {
	val  uint32

	nodes []*node
	points []*Point
}

func (this *node) GobEncode() ([]byte, error) {

    w := new(bytes.Buffer)
/*
	err := binary.Write(w, binary.BigEndian, this.val)
	err = binary.Write(w, binary.BigEndian, this.nodes)
	err = binary.Write(w, binary.BigEndian, this.points)
*/

    encoder := gob.NewEncoder(w)
    err := encoder.Encode(this.val)
    if err != nil { return nil, err }
    err = encoder.Encode(this.nodes)
    if err != nil { return nil, err }
    err = encoder.Encode(this.points)
    if err != nil { return nil, err }

    return w.Bytes(), nil
}

func (this *node) GobDecode(buf []byte) error {
    r := bytes.NewBuffer(buf)
    decoder := gob.NewDecoder(r)
    err := decoder.Decode(&this.val)
    if err != nil { return err }
	err = decoder.Decode(&this.nodes)
    if err != nil { return err }
	
    return nil
}

func insertNode(arr []*node, i uint32, val *node) []*node {

	arr = append(arr, &node{})
	copy(arr[i+1:], arr[i:])
	arr[i] = val

	return arr
}
