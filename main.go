package main

import (
	"io"
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
	//if err := s.ParseAll(); err != nil {
	if err := s.Parse(54865,1); err != nil {
		if err == io.EOF {
			log.Println("End Of File")
			return
		}
		panic(err)
	}
}
