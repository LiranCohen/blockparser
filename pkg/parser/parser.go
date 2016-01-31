package parser

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"time"
)

//Helper Function To Convert VariableInt to Int
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

func (b *Block) Hash() []byte {
	var d []uint8
	//d = append(d, b.magicid[:]...)
	//d = append(d, b.blocklength[:]...)
	d = append(d, b.versionnumber[:]...)
	d = append(d, b.previoushash[:]...)
	d = append(d, b.merkleroot[:]...)
	d = append(d, b.timestamp[:]...)
	d = append(d, b.targetdifficulty[:]...)
	d = append(d, b.nonce[:]...)
	h := sha256.New()
	h.Reset()
	if _, err := h.Write(d); err != nil {
		return []byte{}
	}
	tmp := h.Sum(nil)
	h.Reset()
	if _, err := h.Write(tmp); err != nil {
		return []byte{}
	}
	return h.Sum(nil)
}
func (b *Block) HashString() string {
	var temp []byte
	hash := b.Hash()
	//Not sure how else to converte little endian to string.
	for i := 0; i < len(hash); i++ {
		temp = append([]byte{hash[i]}, temp...)
	}
	return fmt.Sprintf("%x", temp[:])
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
	if b.TimeStamp() > 0 {
		return time.Unix(int64(b.TimeStamp()), 0)
	} else {
		return time.Time{}
	}
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

type Transaction struct {
	versionnumber [4]uint8
	inputcount    []uint8
	Inputs        []TransInput
	outputcount   []uint8
	Outputs       []TransOutput
	locktime      [4]uint8
}

func (t *Transaction) Hash() []uint8 {
	var d []uint8
	d = append(d, t.versionnumber[:]...)
	d = append(d, t.inputcount[:]...)
	for _, ti := range t.Inputs {
		d = append(d, ti.hash[:]...)
		d = append(d, ti.index[:]...)
		d = append(d, ti.scriptlength[:]...)
		d = append(d, ti.script[:]...)
		d = append(d, ti.sequencenumber[:]...)
	}
	d = append(d, t.outputcount[:]...)
	for _, to := range t.Outputs {
		v := make([]byte, 8)
		binary.LittleEndian.PutUint64(v, to.value)
		d = append(d, v[:]...)
		d = append(d, to.scriptlength[:]...)
		d = append(d, to.script[:]...)
	}
	d = append(d, t.locktime[:]...)
	h := sha256.New()
	h.Reset()
	if _, err := h.Write(d); err != nil {
		return []byte{}
	}
	tmp := h.Sum(nil)
	h.Reset()
	if _, err := h.Write(tmp); err != nil {
		return []byte{}
	}
	return h.Sum(nil)
}

func (t *Transaction) VersionNumber() uint32 {
	var v uint32
	reader := bytes.NewReader(t.versionnumber[:])
	if err := binary.Read(reader, binary.LittleEndian, &v); err != nil {
		log.Printf("VersionNumber Error: %v", err)
	}
	return v
}

func (t *Transaction) InputCount() int {
	return VarInt(t.inputcount)
}

func (t *Transaction) OutputCount() int {
	return VarInt(t.outputcount)
}

func (t *Transaction) LockTime() uint32 {
	var v uint32
	reader := bytes.NewReader(t.locktime[:])
	if err := binary.Read(reader, binary.LittleEndian, &v); err != nil {
		log.Printf("TimeStamp Error: %v", err)
	}
	return v
}

func (t *Transaction) LockTimeFormatted() time.Time {
	if t.LockTime() > 0 {
		return time.Unix(int64(t.LockTime()), 0)
	} else {
		return time.Time{}
	}
}

func (t *Transaction) HashString() string {
	var temp []byte
	//Not sure how else to converte little endian to string.
	th := t.Hash()
	for i := 0; i < len(t.Hash()); i++ {
		temp = append([]byte{th[i]}, temp...)
	}
	return fmt.Sprintf("%x", temp[:])
}

type TransInput struct {
	hash           [32]uint8
	index          [4]uint8
	scriptlength   []uint8
	script         []uint8
	sequencenumber [4]uint8
}

func (ti *TransInput) Hash() [32]uint8 {
	//Still in Little Endian
	return ti.hash
}

func (ti *TransInput) HashString() string {
	var temp []byte
	//Not sure how else to converte little endian to string.
	for i := 0; i < cap(ti.hash); i++ {
		temp = append([]byte{ti.hash[i]}, temp...)
	}
	return fmt.Sprintf("%x", temp[:])
}
func (ti *TransInput) Index() uint32 {
	var v uint32
	reader := bytes.NewReader(ti.index[:])
	if err := binary.Read(reader, binary.LittleEndian, &v); err != nil {
		log.Printf("Index Error: %v", err)
	}
	return v
}

func (ti *TransInput) ScriptLength() int {
	return VarInt(ti.scriptlength)
}

func (ti *TransInput) Script() []uint8 {
	return ti.script
}

func (ti *TransInput) SequenceNumber() uint32 {
	var v uint32
	reader := bytes.NewReader(ti.sequencenumber[:])
	if err := binary.Read(reader, binary.LittleEndian, &v); err != nil {
		log.Printf("Sequence Number Error: %v", err)
	}
	return v
}

type TransOutput struct {
	value        uint64
	scriptlength []uint8
	script       []uint8
}

func (to *TransOutput) Value() uint64 {
	return to.value
}

func (to *TransOutput) ScriptLength() int {
	return VarInt(to.scriptlength)
}

func (to *TransOutput) Script() []uint8 {
	return to.script
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

	if err := binary.Read(w, binary.LittleEndian, &block.blocklength); err != nil {
		return &block, err
	}
	if err := binary.Read(w, binary.LittleEndian, &block.versionnumber); err != nil {
		return &block, err
	}
	if err := binary.Read(w, binary.LittleEndian, &block.previoushash); err != nil {
		return &block, err
	}

	if err := binary.Read(w, binary.LittleEndian, &block.merkleroot); err != nil {
		return &block, err
	}

	if err := binary.Read(w, binary.LittleEndian, &block.timestamp); err != nil {
		return &block, err
	}

	if err := binary.Read(w, binary.LittleEndian, &block.targetdifficulty); err != nil {
		return &block, err
	}

	if err := binary.Read(w, binary.LittleEndian, &block.nonce); err != nil {
		return &block, err
	}

	log.Printf("Magic ID: %x\n", block.MagicID())
	log.Printf("Hash: %v\n", block.HashString())
	log.Printf("Block Length: %v\n", block.BlockLength())
	log.Printf("Version Number: %v\n", block.VersionNumber())
	log.Printf("Previous Hash: %v\n", block.PreviousHashString())
	log.Printf("Merkle Root: %v\n", block.MerkleRootString())
	log.Printf("TimeStamp: %v\n", block.TimeStampFormatted())
	log.Printf("Target Difficulty: %v\n", block.TargetDifficulty())
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
	log.Printf("\tTransaction Version Number: %v\n", trans.VersionNumber())

	if b, err := w.ReadByte(); err != nil {
		log.Printf("Input Transaction Count Read Error: %v\n", err)
		return trans, err
	} else {
		var c int
		trans.inputcount = append(trans.inputcount, b)
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
				log.Printf("Transaction Input Count Error: %v\n", err)
			} else {
				trans.inputcount = append(trans.inputcount, b)
			}
		}
	}
	log.Printf("\tInputCount: %v\n", trans.InputCount())
	for i := 0; i < trans.InputCount(); i++ {
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
	log.Printf("\tOutput Count: %v\n", trans.OutputCount())

	for i := 0; i < trans.OutputCount(); i++ {
		if output, err := w.DecodeOutput(); err == nil {
			trans.Outputs = append(trans.Outputs, output)
		} else if err != io.EOF {
			log.Printf("Transaction Parse Error: %v\n", err)
		}
	}

	if err := binary.Read(w, binary.LittleEndian, &trans.locktime); err != nil {
		log.Printf("Transaction Lock Time Error: %v\n", err)
	}
	log.Printf("\tTrsansaction Lock Time: %v\n", trans.LockTimeFormatted())
	log.Printf("\tTransaction Hash: %v\n", trans.HashString())
	return trans, nil
}

func (w *BlockParser) DecodeInput() (TransInput, error) {
	input := TransInput{}
	if err := binary.Read(w, binary.LittleEndian, &input.hash); err != nil {
		log.Printf("Input Hash Error: %v\n", err)
		return input, err
	}
	log.Printf("\tInput Hash: %v\n", input.HashString())
	if err := binary.Read(w, binary.LittleEndian, &input.index); err != nil {
		log.Printf("Input Index Error: %v\n", err)
		return input, err
	}
	log.Printf("\tInput Index: %v\n", input.Index())

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
		log.Printf("\tScript Length: %v\n", input.ScriptLength())

		for i := 0; i < input.ScriptLength(); i++ {
			if b, err := w.ReadByte(); err == nil {
				input.script = append(input.script, b)
			}
		}
		log.Printf("\tScript: %v\n", input.Script())
	}

	if err := binary.Read(w, binary.LittleEndian, &input.sequencenumber); err != nil {
		log.Printf("Input Index Error: %v\n", err)
		return input, err
	}
	log.Printf("\tSequence Number: %v\n", input.SequenceNumber())

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
	log.Printf("\tScript Length: %v\n", VarInt(out.scriptlength))

	for i := 0; i < VarInt(out.scriptlength); i++ {
		if b, err := w.ReadByte(); err == nil {
			out.script = append(out.script, b)
		}
	}
	log.Printf("\tScript: %v\n", out.script)

	return out, nil
}
