package parser

import (
	"io"
	"log"
	"sync"
	"bufio"
	"bytes"
	"encoding/binary"
)

type Stream struct {
	wg     sync.WaitGroup
	stream *bufio.Reader
}

func New(r io.Reader) *Stream {
	return &Stream{
		stream: bufio.NewReader(r),
	}
}

func (s *Stream) ParseBlock(n int, b []byte) error {
	defer s.wg.Done()

	buf := bytes.NewReader(b)
	parser := NewBlockParser(buf, &s.wg)
	_, err := parser.Decode()
	if err != nil {
		return err
	}
	return nil
}

func (s *Stream) ParseLEUint32(b []byte) uint32 {
	var r uint32
	buf := bytes.NewReader(b)
	binary.Read(buf, binary.LittleEndian, &r)
	return r
}

func (s *Stream) Parse(offset, length int) error {

	//Open BlockChain File and load it into a buffer
	reader := s.stream
	blocks := 0
	parsed := 0
	for {
		if length > 0 && parsed == length {
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
		if s.ParseLEUint32(m) == 3652501241 {
			var stmp []byte
			for i := 0; i < 4; i++ {
				b, err := reader.ReadByte()
				if err != nil {
					break
				}
				stmp = append(stmp, b)
			}
			//Temporary place for block in memory
			var block []byte
			block = append(block, m...)
			block = append(block, stmp...)
			//Parse Size of Block and retrieve it
			size := s.ParseLEUint32(stmp)
			for i := 0; i < (int(size)); i++ {
				b, err := reader.ReadByte()
				if err != nil {
					break
				}
				block = append(block, b)
			}
			blocks++
			if offset > 0 &&blocks >= offset {
				//Send the retrieved block to the parser
				s.wg.Add(1)
				//log.Println("FOUND")
				go s.ParseBlock(blocks, block)
				parsed++
			}
		} else {
			log.Printf("Invalid MagicID Looking for next block")
		}
	}

	s.wg.Wait()
	return nil
}
