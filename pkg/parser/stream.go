package parser

import (
	"os"
	"strings"
	"strconv"
	"fmt"
	"io"
	"log"
	"sync"
	"bufio"
	"bytes"
	"errors"

	"github.com/lirancohen/blockparser/pkg/utils"

)


var magic_id = []byte{249,190,180,217}

var ErrEOF = errors.New("EOF")
var ErrNotFound = errors.New("NotFound")


type Stream struct {
	Filename string
	Floor, Ceiling int
	Stream *bufio.Reader
	wg     sync.WaitGroup
}

func New(r io.Reader) *Stream {
	return &Stream{
		Stream: bufio.NewReader(r),
	}
}

func NewStream(name string,r io.Reader, f,c int) *Stream {
	return &Stream{
		Filename: name,
		Stream: bufio.NewReader(r),
		Floor: f,
		Ceiling: c,
	}
}

func EmptyStream() *Stream {
	return &Stream{}
}

func (s *Stream) ParseBlock(n int, b []byte) (*Block, error) {
	defer s.wg.Done()
	buf := bytes.NewReader(b)
	parser := NewBlockParser(buf, &s.wg)
	return parser.Decode(n)
}

func (s *Stream) Next() (*Stream, error){
	_, err := os.Stat("./data/chunks")
	if err != nil{
		return EmptyStream(), err
	}

	dir, err := os.Open("./data/chunks")
	if err != nil {
		return EmptyStream(), err
	}
	defer dir.Close()

	files, err := dir.Readdir(-1)
	if err != nil {
		return EmptyStream(), err
	}

	for _, file := range files {
		st := strings.Split(file.Name(),"_")
		if len(st) >= 1 {
			s2 := strings.Split(st[1], ".")
			if len(s2) > 0 {
				var b, t int
				b, err = strconv.Atoi(st[0])
				t, err = strconv.Atoi(s2[0])
				if err != nil {
					return EmptyStream(), errors.New(fmt.Sprintf("Couldn't read file: %v\n", file.Name()))
				}
				if s.Ceiling == b-1 || (s.Ceiling == 0 && s.Ceiling == b) {
					chunk, err := os.Open("./data/chunks/" + file.Name())
					if err != nil {
						return EmptyStream(), err
					}
					return NewStream(file.Name(),chunk,b,t), nil
				}
			}
		}
	}

	return EmptyStream(), ErrEOF
}

func (s *Stream) Previous()(*Stream, error) {
	_, err := os.Stat("./data/chunks")
	if err != nil{
		return EmptyStream(),err
	}

	dir, err := os.Open("./data/chunks")
	if err != nil {
		return EmptyStream(),err
	}
	defer dir.Close()

	files, err := dir.Readdir(-1)
	if err != nil {
		return EmptyStream(),err
	}

	for _, file := range files {
		st := strings.Split(file.Name(),"_")
		if len(st) >= 1 {
			s2 := strings.Split(st[1], ".")
			if len(s2) > 0 {
				var b, t int
				b, err = strconv.Atoi(st[0])
				t, err = strconv.Atoi(s2[0])
				if err != nil {
					return EmptyStream(),
						errors.New(fmt.Sprintf("Couldn't read file: %v\n", file.Name()))
				}
				if s.Floor == t+1 {
					chunk, err := os.Open("./data/chunks/" + file.Name())
					if err != nil {
						return EmptyStream(), err
					}
					return NewStream(file.Name(),chunk,b,t),nil
				}
			}
		}
	}

	return EmptyStream(), errors.New("this is the first block")
}

func(s *Stream)SeekTransaction(h string) (Transaction, error){
	reader := s.Stream
	blocks := 0
	for {
		var m []byte
		if b, err := reader.ReadByte(); err != nil && err != io.EOF{
			log.Printf("Reader Error On Block: %v\n", blocks)
			log.Printf("Error: %v\n", err)
			break
		} else if err == io.EOF {
			break
		} else {
			//Check to make sure the first byte in the magic ID chunk is 0xF9
			//Remember it's encoded in Little Endian so the first byte will end up being the last when encoded
			if bytes.Compare([]byte{uint8(b)}, []byte{uint8(249)}) == 0 {
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
				continue
			}
		}
		if bytes.Compare(m,magic_id) == 0 {
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
			size := utils.ParseLEUint32(stmp)

			for i := 0; i < (int(size)); i++ {
				b, err := reader.ReadByte()
				if err != nil {
					break
				}
				block = append(block, b)
			}
			s.wg.Add(1)
			fblock, err := s.ParseBlock(blocks + s.Floor, block)
			if err != nil && err.Error() != "unexpected EOF" {
				return Transaction{}, errors.New(fmt.Sprintf("Could not parse block %v:%v\n",blocks, err.Error()))
			}
			for _, t := range fblock.Transactions {
				if h == t.HashString(){
					return t, nil
				}
			}
			blocks++
		} else {
			s.wg.Wait()
			return Transaction{}, errors.New("Invlaid MAGICID, bad block")
		}
	}
	s.wg.Wait()
	return Transaction{}, ErrNotFound
}

func (s *Stream)SeekBlock(n int) (*Block,error) {
	if n < s.Floor || n > s.Ceiling {
		var err error
		s, err = SeekChunk(n)
		if err != nil {
			s.wg.Wait()
			return &Block{}, err
		}
	}

	reader := s.Stream
	target := n - s.Floor
	blocks := 0
	for {
		var m []byte
		if b, err := reader.ReadByte(); err != nil {
			log.Printf("Reader Error On Block: %v\n", blocks)
			log.Printf("Error: %v\n", err)
			break
		} else {
			//Check to make sure the first byte in the magic ID chunk is 0xF9
			//Remember it's encoded in Little Endian so the first byte will end up being the last when encoded
			if bytes.Compare([]byte{uint8(b)}, []byte{uint8(249)}) == 0 {
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
				continue
			}
		}
		if bytes.Compare(m,magic_id) == 0 {
		//if utils.ParseLEUint32(m) == 3652501241 {
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
				size := utils.ParseLEUint32(stmp)

				for i := 0; i < (int(size)); i++ {
					b, err := reader.ReadByte()
					if err != nil {
						break
					}
					block = append(block, b)
				}

			if blocks == target {
				s.wg.Add(1)

				return s.ParseBlock(n, block)

			}
			blocks++
		} else {
			s.wg.Wait()
			return &Block{}, errors.New("Invlaid MAGICID, bad block")
		}
	}
	s.wg.Wait()
	return &Block{}, errors.New("Unknown Error")
}

func (s *Stream) Parse(offset, length int) (int,error) {
	return 0,nil

	//Open BlockChain File and load it into a buffer
	reader := s.Stream
	blocks := 0
	parsed := 0
	for {
		if length > 0 && parsed >= length {
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
		if utils.ParseLEUint32(m) == 3652501241 {
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
			size := utils.ParseLEUint32(stmp)

			for i := 0; i < (int(size)); i++ {
				b, err := reader.ReadByte()
				if err != nil {
					break
				}
				block = append(block, b)
			}

			if offset > 0 && blocks >= offset {
				//Send the retrieved block to the parser
				s.wg.Add(1)
				//log.Println("FOUND")
				parsed++
				go s.ParseBlock(blocks, block)
			}
			blocks++

		} else {
			log.Printf("Invalid MagicID Looking for next block")
		}
	}

	s.wg.Wait()
	return parsed, nil
}

func SeekChunk(n int) (*Stream, error) {
	_, err := os.Stat("./data/chunks")
	if err != nil{
		return EmptyStream(),err
	}

	dir, err := os.Open("./data/chunks")
	if err != nil {
		return EmptyStream(),err
	}
	defer dir.Close()

	files, err := dir.Readdir(-1)
	if err != nil {
		return EmptyStream(),err
	}

	for _, file := range files {
		s := strings.Split(file.Name(),"_")
		if len(s) >= 1 {
			s2 := strings.Split(s[1], ".")
			if len(s2) > 0 {
				var b, t int
				b, err = strconv.Atoi(s[0])
				t, err = strconv.Atoi(s2[0])
				if err != nil {
					return EmptyStream(),
						errors.New(fmt.Sprintf("Couldn't read file: %v\n", file.Name()))
				}
				if n >= b && n <= t {
					chunk, err := os.Open("./data/chunks/" + file.Name())
					if err != nil {
						return EmptyStream(), err
					}
					return NewStream(file.Name(),chunk,b,t),nil
				}
			}
		}
	}

	return EmptyStream(), errors.New("block doesn't exist")
}
