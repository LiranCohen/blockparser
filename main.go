package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"log"
	"os"
	//"strconv"
	"sync"
)

var wg sync.WaitGroup

func main() {
	defer wg.Wait()
	f, err := os.Open("./data/bootstrap.dat")
	if err != nil {
		panic(err)
	}

	reader := bufio.NewReader(f)
	blocks := 0
	for {
		blocks++
		log.Printf("Block #%v\n", blocks)
		if blocks > 10100 {
			break
		}
		var m []byte
		for i := 0; i < 4; i++ {
			b, err := reader.ReadByte()
			if err != nil {
				break
			}
			m = append(m, b)
		}
		var s []byte
		for i := 0; i < 4; i++ {
			b, err := reader.ReadByte()
			if err != nil {
				break
			}
			s = append(s, b)
		}
		var block []byte
		block = append(block, m...)
		block = append(block, s...)
		size := ParseLEUint(s)
		for i := 0; i < (int(size)); i++ {
			b, err := reader.ReadByte()
			if err != nil {
				break
			}
			block = append(block, b)
		}
		if blocks < 10000 {
			continue
		}
		wg.Add(1)
		go ParseBlock(blocks, block)
	}
	log.Printf("Finished")
}

func ParseBlock(n int, b []byte) {
	defer wg.Done()
	log.Printf("Parsing Block #%v\n", n)
	buf := bytes.NewReader(b)
	parser := NewBlockParser(buf)
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

//BlockData Structure
type Block struct {
	MagicID             uint32
	BlockLength         uint32
	VersionNumber       uint32
	PreviousHash        []uint8
	MerkleRoot          []uint8
	TimeStamp           uint32
	TargetDifficulty    uint32
	Nonce               uint32
	TransactionCountRaw []uint8
	TransactionCount    int
	Transactions        []Transaction
}

type Transaction struct {
	VersionNumber  uint32
	InputCountRaw  []uint8
	InputCount     int
	Inputs         []TransInput
	OutputCountRaw []uint8
	OutputCount    int
	Outputs        []TransOutput
	LockTime       uint32
}

type TransInput struct {
	Hash            []uint8
	Index           uint32
	ScriptLengthRaw []uint8
	ScriptLength    int
	Script          []uint8
	SequenceNumber  uint32
}

type TransOutput struct {
	Value           uint64
	ScriptLengthRaw []uint8
	ScriptLength    int
	Script          []uint8
}

type BlockParser struct {
	*bufio.Reader
}

func NewBlockParser(r io.Reader) *BlockParser {
	return &BlockParser{
		Reader: bufio.NewReader(r),
	}
}

func (w *BlockParser) Decode() (*Block, error) {
	block := Block{}

	if err := binary.Read(w, binary.LittleEndian, &block.MagicID); err != nil {
		return &block, err
	}
	log.Printf("Magic ID: %#v\n", block.MagicID)

	if err := binary.Read(w, binary.LittleEndian, &block.BlockLength); err != nil {
		return &block, err
	}
	log.Printf("Block Length: %v\n", block.BlockLength)

	if err := binary.Read(w, binary.LittleEndian, &block.VersionNumber); err != nil {
		return &block, err
	}
	log.Printf("Version Number: %v\n", block.VersionNumber)
	for i := 0; i < 32; i++ {
		b, err := w.ReadByte()
		if err == nil {
			block.PreviousHash = append([]uint8{b}, block.PreviousHash...)
		}
	}
	log.Printf("Previous Hash: %x\n", block.PreviousHash)
	//var s string
	//r := bytes.NewReader(block.PreviousHash)
	//if err := binary.Read(r, binary.LittleEndian, &s); err == nil {
	//log.Printf("Previous Hash: %v\n", s)
	//} else {
	//log.Printf("Hash Error: %v\n", err)
	//}
	for i := 0; i < 32; i++ {
		b, err := w.ReadByte()
		if err == nil {
			block.MerkleRoot = append([]uint8{b}, block.MerkleRoot...)
		}
	}

	if err := binary.Read(w, binary.LittleEndian, &block.TimeStamp); err != nil {
		return &block, err
	}

	if err := binary.Read(w, binary.LittleEndian, &block.TargetDifficulty); err != nil {
		return &block, err
	}
	if err := binary.Read(w, binary.LittleEndian, &block.Nonce); err != nil {
		return &block, err
	}
	log.Printf("Nonce: %v\n", block.Nonce)
	if b, err := w.ReadByte(); err == nil {
		block.TransactionCountRaw = append([]uint8{b}, block.TransactionCountRaw...)
		c := -1
		if uint8(b) == 253 {
			c = 2
		} else if uint8(b) == 254 {
			c = 4
		} else if uint8(b) == 255 {
			c = 8
		}
		for i := 0; i < c; i++ {
			if b, err := w.ReadByte(); err == nil {
				block.TransactionCountRaw = append([]uint8{b}, block.TransactionCountRaw...)
			} else {
				log.Printf("Transaction Count Error: %v\n", err)
				break
			}
		}
	} else {
		log.Printf("Transaction Count Error: %v\n", err)
	}
	switch len(block.TransactionCountRaw) {
	case 1:
		block.TransactionCount = int(block.TransactionCountRaw[0])
	case 3:
		r := bytes.NewReader(block.TransactionCountRaw[:2])
		var i uint16
		if err := binary.Read(r, binary.BigEndian, &i); err != nil {
			log.Printf("Transaction Count Error: %v\n", err)
			break
		}
		block.TransactionCount = int(i)
	case 5:
		r := bytes.NewReader(block.TransactionCountRaw[:4])
		var i uint32
		if err := binary.Read(r, binary.BigEndian, &i); err != nil {
			log.Printf("Transaction Count Error: %v\n", err)
			break
		}
		block.TransactionCount = int(i)
	case 9:
		r := bytes.NewReader(block.TransactionCountRaw[:8])
		var i uint64
		if err := binary.Read(r, binary.BigEndian, &i); err != nil {
			log.Printf("Transaction Count Error: %v\n", err)
			break
		}
		block.TransactionCount = int(i)
	}
	log.Printf("TransactionCount: %v\n", block.TransactionCount)

	for i := 0; i < block.TransactionCount; i++ {
		tp := NewTransParser(w)
		if t, err := tp.Decode(); err == nil {
			log.Println("Adding Transaction")
			block.Transactions = append(block.Transactions, *t)
		} else {
			log.Printf("Transaction Parse Error: %v\n", err)
		}
	}
	//if err := binary.Read(w, binary.LittleEndian, &block.Transactions); err != nil {
	//return &block, err
	//}

	//log.Printf("Block: %#v\n", block)

	return &block, nil
}

type TransParser struct {
	*bufio.Reader
}

func NewTransParser(r io.Reader) *TransParser {
	return &TransParser{
		Reader: bufio.NewReader(r),
	}
}

func (w *TransParser) Decode() (*Transaction, error) {
	block := Transaction{}

	if err := binary.Read(w, binary.LittleEndian, &block.InputCount); err != nil {
		return &block, err
	}

	log.Printf("Block: %#v\n", block)

	return &block, nil
}
