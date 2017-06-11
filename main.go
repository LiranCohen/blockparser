package main

import (
	"io"
	"time"
	"log"
	"os"

	"github.com/lirancohen/blockparser/pkg/parser"
	"strconv"
)

const(
	DEFAULT_LENGTH = 0
	DEFAULT_OFFSET = 0
)

func main() {
	startTime := time.Now()

	args := os.Args[1:]
	offset := getOffset(args)
	length := getLength(args)

	log.Println("Starting...")
	defer log.Println("Finished!")
	defer log.Printf("Time Elapsed %v\n", time.Since(startTime))

	f, err := os.Open("./data/bootstrap.dat")
	if err != nil {
		panic(err)
	}
	s := parser.New(f)
	if err := s.Parse(offset,length); err != nil {
		if err == io.EOF {
			log.Println("End Of File")
			return
		}
		panic(err)
	}
}

func getOffset(args []string) int {
	if(len(args) >= 1) {
		if n,err := strconv.Atoi(args[0]); err == nil{
			return n
		}
	}
	return DEFAULT_OFFSET
}

func getLength(args []string) int {
	if(len(args) >= 2) {
		if n, err := strconv.Atoi(args[1]); err == nil {
			return n
		}
	}
	return DEFAULT_LENGTH
}
