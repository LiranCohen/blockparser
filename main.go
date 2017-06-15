package main

import (
	//"io"
	"fmt"
	"time"
	"path/filepath"
	"strconv"
	"log"
	"os"

	"github.com/lirancohen/blockparser/pkg/chunker"
	"github.com/lirancohen/blockparser/pkg/parser"
)

const(
	DEFAULT_LENGTH = 1
	DEFAULT_OFFSET = 0
)

func main() {
	startTime := time.Now()

	defer log.Println("Finished!")

	args := os.Args[1:]

	if ok := getRechunk(args); ok {
		log.Println("Rechunking...")
		if err := bootStrap(true); err != nil {
			panic(err)
		}
		return
	}

	log.Println("Bootstrapping...")
	err := bootStrap(false)
	if err != nil {
		panic(err)
	}
	log.Println("Starting...")

	chunk := parser.EmptyStream()

	seekBlock := time.Now()
	b, err := chunk.SeekBlock(153398)
	if err != nil {
		panic(err)
	}
	log.Println("Found Block in: %v\n", time.Since(seekBlock).String())

	fmt.Println(b.PrintBlockInfo())


	//var offset,length int
	//
	//s := parser.New(f)
	//if o,l,ok := getRange(args); ok{
	//	offset = o
	//	length = l
	//	log.Println("RANGE")
	//} else {
	//	offset = getOffset(args)
	//	length = getLength(args)
	//}
	//
	//if n,err := s.Parse(offset,length); err != nil {
	//	log.Printf("Blocks parsed: %v\n", n)
	//	if err == io.EOF {
	//		log.Println("End Of File")
	//		return
	//	}
	//	panic(err)
	//} else {
	//	log.Printf("Blocks parsed: %v\n", n)
	//}
	log.Printf("Time Elapsed %v\n", time.Since(startTime).String())
}

func getRechunk(args []string) bool {
	if len(args) > 0 && args[0] == "--rechunk" {
		return true
	}
	return false
}

func getRange(args []string) (int, int, bool) {
	if(len(args) >= 3 && args[0] == "--range"){
		i1,err1 := strconv.Atoi(args[1])
		i2,err2 := strconv.Atoi(args[2])
		if(err1 != nil || err2 != nil){
			log.Println("Error: entries must be in integer format")
			return 0,0,false
		}
		offset := i2 - i1
		if (offset <= 0){
			return 0,0,false
		}
		return i1, offset, true
	}
	return 0,0,false
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

func bootStrap(rechunk bool) error {

	_, err := os.Stat("./data/chunks")
	if err != nil{
		if os.IsNotExist(err){
			os.MkdirAll("./data/chunks", os.ModePerm)
		}
	}

	if !rechunk {
		dir, err := os.Open("./data/chunks")
		if err != nil {
			return err
		}
		defer dir.Close()

		files, err := dir.Readdir(-1)
		if err != nil {
			return err
		}

		for _, file := range files {
			if filepath.Ext(file.Name()) == ".current" {
				return nil
			}
		}
	}
	os.RemoveAll("./data/chunks/")
	os.MkdirAll("./data/chunks", os.ModePerm)

	f, err := os.Open("./data/bootstrap.dat")
	if err != nil {
		return err
	}

	c := chunker.New(f)
	n, err := c.Chunk()
	if err != nil {
		return err
	}
	log.Printf("%v Blocks Written", n)

	return nil
}
