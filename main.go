package main

import (
	"log"
	"os"

	"github.com/lirancohen/blockparser/pkg/parser"
)

func main() {
	log.Println("Starting")
	defer log.Println("Finished")
	f, err := os.Open("./data/bootstrap.dat")
	if err != nil {
		panic(err)
	}
	s := parser.New(f)
	if err := s.ParseAll(); err != nil {
		panic(err)
	}
}
