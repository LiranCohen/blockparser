package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"time"
	//"strconv"
	"sync"
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
	magicid          [4]byte
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
	reader := bytes.NewReader(b.magicid[:])
	if err := binary.Read(reader, binary.LittleEndian, &v); err != nil {
		log.Printf("MagicID Error: %v", err)
	}
	return v
}
func (b *Block) BlockLength() uint32 {
	var v uint32
	reader := bytes.NewReader(b.blocklength[:])
	if err := binary.Read(reader, binary.LittleEndian, &v); err != nil {
		log.Printf("BlockLength Error: %v", err)
	}
	return v
}
func (b *Block) VersionNumber() uint32 {
	var v uint32
	reader := bytes.NewReader(b.versionnumber[:])
	if err := binary.Read(reader, binary.LittleEndian, &v); err != nil {
		log.Printf("VersionNumber Error: %v", err)
	}
	return v
}

func (b *Block) PreviousHash() [32]uint8 {
	//Still in Little Endian
	return b.previoushash
}

func (b *Block) PreviousHashString() string {
	var temp []byte
	//Not sure how else to converte little endian to string.
	for i := 0; i < cap(b.previoushash); i++ {
		temp = append([]byte{b.previoushash[i]}, temp...)
	}
	return fmt.Sprintf("%x", temp[:])
}

func (b *Block) MerkleRoot() [32]uint8 {
	//Still in Little Endian
	return b.merkleroot
}

func (b *Block) MerkleRootString() string {
	var temp []byte
	//Not sure how else to converte little endian to string.
	for i := 0; i < cap(b.merkleroot); i++ {
		temp = append([]byte{b.merkleroot[i]}, temp...)
	}
	return fmt.Sprintf("%x", temp[:])
}

func (b *Block) TimeStamp() uint32 {
	var v uint32
	reader := bytes.NewReader(b.timestamp[:])
	if err := binary.Read(reader, binary.LittleEndian, &v); err != nil {
		log.Printf("TimeStamp Error: %v", err)
	}
	return v
}

func (b *Block) TimeStampFormatted() time.Time {
	var t time.Time
	t = time.Unix(1231469665, 0)
	return t
}

func (b *Block) TargetDifficulty() uint32 {
	var v uint32
	reader := bytes.NewReader(b.targetdifficulty[:])
	if err := binary.Read(reader, binary.LittleEndian, &v); err != nil {
		log.Printf("TargetDifficulty Error: %v", err)
	}
	return v

}

func (b *Block) Nonce() uint32 {
	var v uint32
	reader := bytes.NewReader(b.nonce[:])
	if err := binary.Read(reader, binary.LittleEndian, &v); err != nil {
		log.Printf("Nonce Error: %v", err)
	}
	return v
}
func VarInt(input []byte) int {
	var v int
	switch len(input) {
	case 1:
		v = int(input[0])
	case 3:
		r := bytes.NewReader(input[:2])
		var i uint16
		if err := binary.Read(r, binary.BigEndian, &i); err != nil {
			log.Printf("Transaction Count Error: %v\n", err)
			break
		}
		v = int(i)
	case 5:
		r := bytes.NewReader(input[:4])
		var i uint32
		if err := binary.Read(r, binary.BigEndian, &i); err != nil {
			log.Printf("Transaction Count Error: %v\n", err)
			break
		}
		v = int(i)
	case 9:
		r := bytes.NewReader(input[:8])
		var i uint64
		if err := binary.Read(r, binary.BigEndian, &i); err != nil {
			log.Printf("Transaction Count Error: %v\n", err)
			break
		}
		v = int(i)
	}
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

func (t *Transaction) OutputCount() int {
	var v int
	switch len(t.outputcount) {
	case 1:
		v = int(t.outputcount[0])
	case 3:
		r := bytes.NewReader(t.outputcount[:2])
		var i uint16
		if err := binary.Read(r, binary.BigEndian, &i); err != nil {
			log.Printf("Transaction Count Error: %v\n", err)
			break
		}
		v = int(i)
	case 5:
		r := bytes.NewReader(t.outputcount[:4])
		var i uint32
		if err := binary.Read(r, binary.BigEndian, &i); err != nil {
			log.Printf("Transaction Count Error: %v\n", err)
			break
		}
		v = int(i)
	case 9:
		r := bytes.NewReader(t.outputcount[:8])
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
	value        uint64
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
	log.Printf("Magic ID: %x\n", block.MagicID())

	if err := binary.Read(w, binary.LittleEndian, &block.blocklength); err != nil {
		return &block, err
	}
	log.Printf("Block Length: %v\n", block.BlockLength())

	if err := binary.Read(w, binary.LittleEndian, &block.versionnumber); err != nil {
		return &block, err
	}
	log.Printf("Version Number: %v\n", block.VersionNumber())

	if err := binary.Read(w, binary.LittleEndian, &block.previoushash); err != nil {
		return &block, err
	}
	log.Printf("Previous Hash: %v\n", block.PreviousHashString())

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

	log.Printf("Merkle Root: %v\n", block.MerkleRootString())

	//for i := 0; i < 32; i++ {
	//b, err := w.ReadByte()
	//if err == nil {
	//block.MerkleRoot = append([]uint8{b}, block.MerkleRoot...)
	//}
	//}

	if err := binary.Read(w, binary.LittleEndian, &block.timestamp); err != nil {
		return &block, err
	}
	log.Printf("TimeStamp: %v\n", block.TimeStampFormatted())

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
	log.Printf("TransactionCount: %v\n", VarInt(block.transactioncount))

	for i := 0; i < VarInt(block.transactioncount); i++ {
		if t, err := w.DecodeTrans(); err == nil {
			block.Transactions = append(block.Transactions, t)
		} else if err != io.EOF {
			log.Printf("Transaction Parse Error: %v\n", err)
		}
	}
	//log.Printf("Block: %#v\n", block)

	return &block, nil
}

func (w *BlockParser) DecodeTrans() (Transaction, error) {
	trans := Transaction{}
	log.Println("Decoding Transaction")
	if err := binary.Read(w, binary.LittleEndian, &trans.versionnumber); err != nil {
		log.Printf("Transaction Version Error: %v\n", err)
		return trans, err
	}
	log.Printf("\tTransaction VersionNumber: %v\n", trans.versionnumber)

	if b, err := w.ReadByte(); err != nil {
		log.Printf("Input Transaction Count Read Error: %v\n", err)
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
	log.Printf("\tInputCount: %v\n", VarInt(trans.inputcount))
	for i := 0; i < VarInt(trans.inputcount); i++ {
		if input, err := w.DecodeInput(); err == nil {
			trans.Inputs = append(trans.Inputs, input)
		} else if err != io.EOF {
			log.Printf("Transaction Parse Error: %v\n", err)
		}
	}

	if b, err := w.ReadByte(); err != nil {
		log.Printf("Output Transaction Count Read Error: %v\n", err)
		return trans, err
	} else {
		trans.outputcount = append(trans.outputcount, b)
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
				trans.outputcount = append(trans.outputcount, b)
			} else {
				log.Printf("Transaction Output Count Error: %v\n", err)
				break
			}
		}
	}
	log.Printf("Output Count: %v\n", VarInt(trans.outputcount))
	for i := 0; i < VarInt(trans.outputcount); i++ {
		if output, err := w.DecodeOutput(); err == nil {
			trans.Outputs = append(trans.Outputs, output)
		} else if err != io.EOF {
			log.Printf("Transaction Parse Error: %v\n", err)
		}
	}
	if err := binary.Read(w, binary.LittleEndian, &trans.loctime); err != nil {
		log.Printf("Transaction Lock Time Error: %v\n", err)
	}
	log.Printf("Trsansaction Lock Time: %#v\n", trans.loctime)

	return trans, nil
}

func (w *BlockParser) DecodeInput() (TransInput, error) {
	input := TransInput{}
	if err := binary.Read(w, binary.LittleEndian, &input.hash); err != nil {
		log.Printf("Input Hash Error: %v\n", err)
		return input, err
	}
	log.Printf("\tInput Hash: %v\n", input.hash)
	if err := binary.Read(w, binary.LittleEndian, &input.index); err != nil {
		log.Printf("Input Index Error: %v\n", err)
		return input, err
	}
	log.Printf("\tInput Index: %v\n", input.index)

	if b, err := w.ReadByte(); err != nil {
		log.Printf("Script Length Read Error: %v\n", err)
		return input, err
	} else {
		input.scriptlength = append(input.scriptlength, b)
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
				input.scriptlength = append(input.scriptlength, b)
			} else {
				log.Printf("Input Script Length Error: %v\n", err)
				break
			}
		}
		log.Printf("Script Length: %v\n", VarInt(input.scriptlength))

		for i := 0; i < VarInt(input.scriptlength); i++ {
			if b, err := w.ReadByte(); err == nil {
				input.script = append(input.script, b)
			}
		}
		log.Printf("Script: %v\n", input.script)
	}

	if err := binary.Read(w, binary.LittleEndian, &input.sequencenumber); err != nil {
		log.Printf("Input Index Error: %v\n", err)
		return input, err
	}
	log.Printf("\tSequence Number: %v\n", input.sequencenumber)

	return input, nil
}

func (w *BlockParser) DecodeOutput() (TransOutput, error) {
	out := TransOutput{}
	if err := binary.Read(w, binary.LittleEndian, &out.value); err != nil {
		log.Printf("Output Value Error: %v\n", err)
		return out, err
	}
	log.Printf("\tOutput Value: %v\n", out.value)

	if b, err := w.ReadByte(); err != nil {
		log.Printf("Script Length Read Error: %v\n", err)
		return out, err
	} else {
		out.scriptlength = append(out.scriptlength, b)
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
				out.scriptlength = append(out.scriptlength, b)
			} else {
				log.Printf("Input Script Length Error: %v\n", err)
				break
			}
		}
	}
	log.Printf("Script Length: %v\n", VarInt(out.scriptlength))

	for i := 0; i < VarInt(out.scriptlength); i++ {
		if b, err := w.ReadByte(); err == nil {
			out.script = append(out.script, b)
		}
	}
	log.Printf("Script: %v\n", out.script)

	return out, nil
}
