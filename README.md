# pointDB

[![godoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://godoc.org/github.com/obsius/pointDB)
[![coverage](https://coveralls.io/repos/github/obsius/pointDB/badge.svg?branch=master)](https://coveralls.io/github/obsius/pointDB?branch=master)
[![go report](https://goreportcard.com/badge/obsius/pointDB)](https://goreportcard.com/report/obsius/pointDB)
[![build](https://travis-ci.org/obsius/pointDB.svg?branch=master)](https://travis-ci.org/obsius/pointDB)
[![license](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/obsius/pointDB/master/LICENSE)

A fast, lightweight, in-memory database for querying coordinates.  Written in Go.

## Installation
```golang
type ConvertibleTo interface {
	ConvertTo(interface{}) (bool, error)
}
```

## Querying
```golang
type ConvertibleTo interface {
	ConvertTo(interface{}) (bool, error)
}

## Benchmarks
operation|ns/op|# operations|total time
-|-|-|-