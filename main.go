package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	//"strconv"
	"sync"

	"github.com/lirancohen/btc/pkg/parser"
)

var wg sync.WaitGroup

func main() {
	defer wg.Wait()

	//Open BlockChain File and load it into a buffer
	f, err := os.Open("./data/bootstrap.dat")
	if err != nil {
		panic(err)
	}
	reader := bufio.NewReader(f)
	blocks := 0
	parsed := 0
	for {
		//log.Printf("Block #%v\n", blocks)
		if parsed > 1 {
			break
		}
		//Temporary holder for what might be the MagicID
		var m []byte
		if b, err := reader.ReadByte(); err != nil {
			log.Printf("Reader Error On Block: %v\n", blocks)
			log.Printf("Error: %v\n", err)
			break
		} else {
			//Check to make sure the first byte in the magic ID chunk is 0xF9
			//Remember it's encoded in Little Endian so the first byte will end up being the last when encoded
			if uint8(b) == 249 {
				m = append(m, b)
				for i := 0; i < 3; i++ {
					if b, err := reader.ReadByte(); err != nil {
						log.Printf("Reader Error inside MagicID On Block: %v\n", blocks)
						log.Printf("Error: %v\n", err)
						break
					} else {
						m = append(m, b)
					}
				}
			} else {
				log.Printf("Invalid Byte: %v\n", b)
				continue
			}
		}
		//Check If the MagicID == 0xD9B4BEF9
		//If so, this is the beginning of a BitCoin BlockChain Block
		if ParseLEUint(m) == 3652501241 {
			var s []byte
			for i := 0; i < 4; i++ {
				b, err := reader.ReadByte()
				if err != nil {
					break
				}
				s = append(s, b)
			}
			//Temporary place for block in memory
			var block []byte
			block = append(block, m...)
			block = append(block, s...)
			//Parse Size of Block and retrieve it
			size := ParseLEUint(s)
			for i := 0; i < (int(size)); i++ {
				b, err := reader.ReadByte()
				if err != nil {
					break
				}
				block = append(block, b)
			}
			blocks++
			if blocks < 100000 {
				continue
			}
			//Send the retrieved block to the parser
			wg.Add(1)
			go ParseBlock(blocks, block)
			parsed++
		} else {
			log.Printf("Invalid MagicID Looking for next block")
		}
	}
	log.Printf("Finished")
}

func ParseBlock(n int, b []byte) {
	defer wg.Done()
	defer fmt.Println(fmt.Sprintf("###############  End  Block #%v  ###############\n", (n - 1)))
	buf := bytes.NewReader(b)
	parser := parser.NewBlockParser(buf)
	_, err := parser.Decode()
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}
}

func ParseLEUint(b []byte) uint32 {
	var r uint32
	buf := bytes.NewReader(b)
	binary.Read(buf, binary.LittleEndian, &r)
	return r
}
