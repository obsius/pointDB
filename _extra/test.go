package main

import (
	"fmt"
)

type mountain struct {
	val *int
	data interface{}
}

func main() {

	arr := make([]*mountain, 10)
	for i := range arr {
		v := i
		arr[i].val = &v
		arr[i].data = make([]*int, 10)
	}

	pointer := &arr

	fmt.Println(arr)
	fmt.Println(pointer)

	*arr[0] = 33

	**pointer = 33

	for i := range arr {
		fmt.Printf("%v ", *arr[i])
	}
}