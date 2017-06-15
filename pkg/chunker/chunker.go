package chunker

import (
	"bufio"
	"io"
	"fmt"
	"os"
	"log"
	"sync"

	"github.com/lirancohen/blockparser/pkg/utils"
)

const (
	CHUNK_LENGTH = 9999
)


type ChainChunker struct {
	*bufio.Reader
	File *bufio.Writer
	wg sync.WaitGroup
}

func New(r io.Reader) *ChainChunker {
	return &ChainChunker{
		Reader: bufio.NewReader(r),
	}
}

func (c *ChainChunker) Chunk() (int,error) {

	//Open BlockChain File and load it into a buffer
	var path string
	reader := c.Reader
	blocks := 0
	written := 0
	chunkLength := 0
	chunkStart := 0
	for {
		if chunkLength == 0 {
			chunkLength =  CHUNK_LENGTH
		}

		if (blocks == (chunkStart  + chunkLength + 1)) || blocks == 0 || blocks == chunkLength + 1 {
			chunkStart = blocks

			bucket := fmt.Sprintf("%v_%v", chunkStart, (chunkStart + chunkLength))
			var err error
			path = fmt.Sprintf("./data/chunks/%v.dat",bucket)
			w, err := os.Create(path)
			c.File = bufio.NewWriter(w)
			if err != nil {
				return written, err
			}
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
				if err := c.File.WriteByte(b); err != nil {
					log.Printf("Error at byte: %v", b)
				}
				m = append(m, b)
				for i := 0; i < 3; i++ {
					if b, err := reader.ReadByte(); err != nil {
						log.Printf("Reader Error inside MagicID On Block: %v\n", blocks)
						log.Printf("Error: %v\n", err)
						break
					} else {
						if err := c.File.WriteByte(b); err != nil {
							log.Printf("Error at byte: %v", b)
						}
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
		if utils.ParseLEUint32(m) == 3652501241 {
			var stmp []byte
			for i := 0; i < 4; i++ {
				b, err := reader.ReadByte()
				if err != nil {
					break
				}
				if err := c.File.WriteByte(b); err != nil {
					log.Printf("Error at byte: %v", b)
				}
				stmp = append(stmp, b)
			}

			//Parse Size of Block and retrieve it
			size := utils.ParseLEUint32(stmp)

			for i := 0; i < (int(size)); i++ {
				b, err := reader.ReadByte()
				if err != nil {
					break
				}
				if err := c.File.WriteByte(b); err != nil {
					log.Printf("Error at byte: %v", b)
				}
			}

			blocks++
			//write block to current chunk

			written++
		} else {
			log.Printf("Invalid MagicID Looking for next block")
		}
	}

	var pathRenamed string
	if blocks == (chunkStart + chunkLength){
		pathRenamed = fmt.Sprintf("%v.%v",path, "current" )
	} else {
		pathRenamed = fmt.Sprintf("%v.%v.%v",path, blocks,"current" )
	}

	os.Rename(path,pathRenamed)
	return written, nil
}
