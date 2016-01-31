package parser

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"time"
)

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

func (b *Block) TransactionCount() int {
	return VarInt(b.transactioncount)
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

type Transaction struct {
	versionnumber [4]uint8
	inputcount    []uint8
	Inputs        []TransInput
	outputcount   []uint8
	Outputs       []TransOutput
	loctime       [4]uint8
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

	if err := binary.Read(w, binary.LittleEndian, &block.merkleroot); err != nil {
		return &block, err
	}

	log.Printf("Merkle Root: %v\n", block.MerkleRootString())

	if err := binary.Read(w, binary.LittleEndian, &block.timestamp); err != nil {
		return &block, err
	}
	log.Printf("TimeStamp: %v\n", block.TimeStampFormatted())

	if err := binary.Read(w, binary.LittleEndian, &block.targetdifficulty); err != nil {
		return &block, err
	}
	log.Printf("Target Difficulty: %v\n", block.TargetDifficulty())

	if err := binary.Read(w, binary.LittleEndian, &block.nonce); err != nil {
		return &block, err
	}
	log.Printf("Nonce: %v\n", block.Nonce())

	if b, err := w.ReadByte(); err != nil {
		log.Printf("Transaction Count Read Error: %v\n", err)
	} else {
		var c int
		block.transactioncount = append(block.transactioncount, b)
		if uint8(b) == 253 {
			c = 2
		} else if uint8(b) == 254 {
			c = 4
		} else if uint8(b) == 255 {
			c = 8
		}
		if c > 1 {
			d := make([]uint8, c)
			if err := binary.Read(w, binary.LittleEndian, &d); err != nil {
				log.Printf("Transaction Count Error: %v\n", err)
			} else {
				block.transactioncount = append(block.transactioncount, d...)
			}
		}
	}
	log.Printf("TransactionCount: %v\n", block.TransactionCount())

	for i := 0; i < block.TransactionCount(); i++ {
		if t, err := w.DecodeTrans(); err == nil {
			block.Transactions = append(block.Transactions, t)
		} else if err != io.EOF {
			log.Printf("Transaction Parse Error: %v\n", err)
		}
	}

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
		var c int
		trans.outputcount = append(trans.outputcount, b)
		if uint8(b) == 253 {
			c = 2
		} else if uint8(b) == 254 {
			c = 4
		} else if uint8(b) == 255 {
			c = 8
		}
		if c > 1 {
			d := make([]uint8, c)
			if err := binary.Read(w, binary.LittleEndian, &d); err != nil {
				log.Printf("Transaction Output Count Error: %v\n", err)
			} else {
				trans.outputcount = append(trans.outputcount, d...)
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
		var c int
		input.scriptlength = append(input.scriptlength, b)
		if uint8(b) == 253 {
			c = 2
		} else if uint8(b) == 254 {
			c = 4
		} else if uint8(b) == 255 {
			c = 8
		}
		if c > 1 {
			d := make([]uint8, c)
			if err := binary.Read(w, binary.LittleEndian, &d); err != nil {
				log.Printf("Input Script Length Error: %v\n", err)
			} else {
				input.scriptlength = append(input.scriptlength, d...)
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
		var c int
		out.scriptlength = append(out.scriptlength, b)
		if uint8(b) == 253 {
			c = 2
		} else if uint8(b) == 254 {
			c = 4
		} else if uint8(b) == 255 {
			c = 8
		}
		if c > 1 {
			d := make([]uint8, c)
			if err := binary.Read(w, binary.LittleEndian, &d); err != nil {
				log.Printf("Input Script Length Error: %v\n", err)
			} else {
				out.scriptlength = append(out.scriptlength, d...)
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
