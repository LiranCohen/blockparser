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
		log.Printf("Block #%v\n", blocks)
		if blocks > 100100 {
			break
		}
		var m []byte
		if b, err := reader.ReadByte(); err != nil {
			log.Printf("Reader Error On Block: %v\n", blocks)
			log.Printf("Error: %v\n", err)
			break
		} else {
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
		if ParseLEUint(m) == 3652501241 {
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
			blocks++
			if blocks < 100000 {
				continue
			}
			wg.Add(1)
			go ParseBlock(blocks, block)
		} else {
			log.Printf("Invalid MagicID Looking for next block")
		}
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
	magicid          [4]uint8
	blocklength      [4]uint8
	versionnumber    [4]uint8
	previoushash     [32]uint8
	merkleroot       [32]uint8
	timestamp        [4]uint8
	targetdifficulty [4]uint8
	nonce            [4]uint8
	transactioncount []uint8
	Transactions     []Transaction
}

func (b *Block) MagicID() uint32 {
	var v uint32
	return v
}
func (b *Block) BlockLength() uint32 {
	var v uint32
	return v
}
func (b *Block) VersionNumber() uint32 {
	var v uint32
	return v
}
func (b *Block) PreviousHash() [32]uint8 {
	var v [32]uint8
	return v
}
func (b *Block) PreviousHashString() string {
	var v string
	return v
}
func (b *Block) MerkleRoot() [32]uint8 {
	var v [32]uint8
	return v
}
func (b *Block) MerkleRootString() string {
	var v string
	return v
}
func (b *Block) TimeStamp() uint32 {
	var v uint32
	return v
}
func (b *Block) TargetDifficulty() uint32 {
	var v uint32
	return v

}
func (b *Block) Nonce() uint32 {
	var v uint32
	return v
}
func (b *Block) TransactionCount() int {
	var v int
	switch len(b.transactioncount) {
	case 1:
		v = int(b.transactioncount[0])
	case 3:
		r := bytes.NewReader(b.transactioncount[:2])
		var i uint16
		if err := binary.Read(r, binary.BigEndian, &i); err != nil {
			log.Printf("Transaction Count Error: %v\n", err)
			break
		}
		v = int(i)
	case 5:
		r := bytes.NewReader(b.transactioncount[:4])
		var i uint32
		if err := binary.Read(r, binary.BigEndian, &i); err != nil {
			log.Printf("Transaction Count Error: %v\n", err)
			break
		}
		v = int(i)
	case 9:
		r := bytes.NewReader(b.transactioncount[:8])
		var i uint64
		if err := binary.Read(r, binary.BigEndian, &i); err != nil {
			log.Printf("Transaction Count Error: %v\n", err)
			break
		}
		v = int(i)
	}
	return v
}

type Transaction struct {
	versionnumber [4]uint8
	inputcount    []uint8
	Inputs        []TransInput
	outputcount   []uint8
	Outputs       []TransOutput
	loctime       [4]uint8
}

func (t *Transaction) InputCount() int {
	var v int
	switch len(t.inputcount) {
	case 1:
		v = int(t.inputcount[0])
	case 3:
		r := bytes.NewReader(t.inputcount[:2])
		var i uint16
		if err := binary.Read(r, binary.BigEndian, &i); err != nil {
			log.Printf("Transaction Count Error: %v\n", err)
			break
		}
		v = int(i)
	case 5:
		r := bytes.NewReader(t.inputcount[:4])
		var i uint32
		if err := binary.Read(r, binary.BigEndian, &i); err != nil {
			log.Printf("Transaction Count Error: %v\n", err)
			break
		}
		v = int(i)
	case 9:
		r := bytes.NewReader(t.inputcount[:8])
		var i uint64
		if err := binary.Read(r, binary.BigEndian, &i); err != nil {
			log.Printf("Transaction Count Error: %v\n", err)
			break
		}
		v = int(i)
	}
	return v
}

type TransInput struct {
	hash           [32]uint8
	index          [4]uint8
	scriptlength   []uint8
	script         []uint8
	sequencenumber [4]uint8
}

type TransOutput struct {
	value        [8]uint64
	scriptlength []uint8
	script       []uint8
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
	if err := binary.Read(w, binary.LittleEndian, &block.magicid); err != nil {
		return &block, err
	}
	log.Printf("Magic ID: %#v\n", block.magicid)

	if err := binary.Read(w, binary.LittleEndian, &block.blocklength); err != nil {
		return &block, err
	}
	log.Printf("Block Length: %v\n", block.BlockLength)

	if err := binary.Read(w, binary.LittleEndian, &block.versionnumber); err != nil {
		return &block, err
	}
	log.Printf("Version Number: %v\n", block.versionnumber)

	if err := binary.Read(w, binary.LittleEndian, &block.previoushash); err != nil {
		return &block, err
	}
	log.Printf("Previous Hash: %v\n", block.previoushash)

	//for i := 0; i < 32; i++ {
	//b, err := w.ReadByte()
	//if err == nil {
	//block.PreviousHash = append([]uint8{b}, block.PreviousHash...)
	//}
	//}
	//log.Printf("Previous Hash: %x\n", block.PreviousHash)
	//var s string
	//r := bytes.NewReader(block.PreviousHash)
	//if err := binary.Read(r, binary.LittleEndian, &s); err == nil {
	//log.Printf("Previous Hash: %v\n", s)
	//} else {
	//log.Printf("Hash Error: %v\n", err)
	//}

	if err := binary.Read(w, binary.LittleEndian, &block.merkleroot); err != nil {
		return &block, err
	}
	log.Printf("Merkle Root: %v\n", block.merkleroot)

	//for i := 0; i < 32; i++ {
	//b, err := w.ReadByte()
	//if err == nil {
	//block.MerkleRoot = append([]uint8{b}, block.MerkleRoot...)
	//}
	//}

	if err := binary.Read(w, binary.LittleEndian, &block.timestamp); err != nil {
		return &block, err
	}
	log.Printf("TimeStamp: %v\n", block.timestamp)

	if err := binary.Read(w, binary.LittleEndian, &block.targetdifficulty); err != nil {
		return &block, err
	}
	log.Printf("Target Difficulty: %v\n", block.targetdifficulty)

	if err := binary.Read(w, binary.LittleEndian, &block.nonce); err != nil {
		return &block, err
	}
	log.Printf("Nonce: %v\n", block.nonce)

	if b, err := w.ReadByte(); err != nil {
		log.Printf("Transaction Count Read Error: %v\n", err)
	} else {
		block.transactioncount = append(block.transactioncount, b)
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
				block.transactioncount = append(block.transactioncount, b)
			} else {
				log.Printf("Transaction Count Error: %v\n", err)
				break
			}
		}
	}
	log.Printf("TransactionCount: %v\n", block.TransactionCount())

	for i := 0; i < block.TransactionCount(); i++ {
		tp := NewTransParser(w)
		if t, err := tp.Decode(); err == nil {
			block.Transactions = append(block.Transactions, t)
		} else if err != io.EOF {
			log.Printf("Transaction Parse Error: %v\n", err)
		}
	}
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

func (w *TransParser) Decode() (Transaction, error) {
	trans := Transaction{}
	log.Println("Decoding Transaction")

	if err := binary.Read(w, binary.LittleEndian, &trans.versionnumber); err != nil {
		log.Printf("Transaction Version Error: %v\n", err)
		return trans, err
	}
	log.Printf("\tTransaction VersionNumber: %v\n", trans.versionnumber)

	if b, err := w.ReadByte(); err != nil {
		log.Printf("Transaction Count Read Error: %v\n", err)
		return trans, err
	} else {
		trans.inputcount = append(trans.inputcount, b)
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
				trans.inputcount = append(trans.inputcount, b)
			} else {
				log.Printf("Transaction Input Count Error: %v\n", err)
				break
			}
		}
	}
	log.Printf("InputCount: %v\n", trans.InputCount())

	return trans, nil
}
