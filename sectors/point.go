package sectors

import (
	"encoding/gob"
	_"encoding/binary"
	"bytes"
	_"fmt"
)

type Point struct {
	ID uint32
	Coord Coord
}

func (this *Point) GobEncode() ([]byte, error) {

    w := new(bytes.Buffer)

 //   err := binary.Write(w, binary.BigEndian, this.ID)
//	err = binary.Write(w, binary.BigEndian, this.Coord)
    encoder := gob.NewEncoder(w)
    err := encoder.Encode(this.ID)
    if err != nil { return nil, err }
    err = encoder.Encode(this.Coord)
    if err != nil { return nil, err }
//fmt.Print(w.Bytes())
    return w.Bytes(), nil
}

func (this *Point) GobDecode(buf []byte) error {
  //  r := bytes.NewBuffer(buf)
  //  decoder := gob.NewDecoder(r)
   // err := decoder.Decode(&this.val)
  //  if err != nil { return err }
	//err = decoder.Decode(&this.data)
 //   if err != nil { return err }
	
    return nil
}