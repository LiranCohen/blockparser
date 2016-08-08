package parser

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
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

func New(r io.Reader) *Stream {
	s := Stream{}
	var err error
	s.db, err = gorm.Open("sqlite3", "./blockchain.db")
	if err != nil {
		panic(err)
	}
	s.stream = bufio.NewReader(r)
	//s.db.LogMode(true)
	//s.db.CreateTable(&Block{})
	//s.db.CreateTable(&Transaction{})
	//s.db.CreateTable(&TransInput{})
	//s.db.CreateTable(&TransOutput{})
	return &s
}

type Stream struct {
	wg     sync.WaitGroup
	db     *gorm.DB
	stream *bufio.Reader
}

func (s *Stream) ParseBlock(n int, b []byte) {
	defer s.wg.Done()
	//defer fmt.Println(fmt.Sprintf("###############  End  Block #%v  ###############\n", (n - 1)))
	buf := bytes.NewReader(b)
	parser := NewBlockParser(buf)
	_, err := parser.Decode()
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}
}

func (s *Stream) ParseLEUint32(b []byte) uint32 {
	var r uint32
	buf := bytes.NewReader(b)
	binary.Read(buf, binary.LittleEndian, &r)
	return r
}

func (s *Stream) ParseBlockN(n int) error {
	defer s.wg.Wait()

	//Open BlockChain File and load it into a buffer
	reader := s.stream
	blocks := 0
	parsed := 0
	for {
		//log.Printf("Block #%v\n", blocks)
		if parsed == 1 {
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
			if blocks-1 < n {
				continue
			}
			//Send the retrieved block to the parser
			s.wg.Add(1)
			go s.ParseBlock(blocks, block)
			parsed++
		} else {
			log.Printf("Invalid MagicID Looking for next block")
		}
	}

	return nil

}

//Does not currently parse all
func (s *Stream) ParseAll() error {
	defer s.wg.Wait()

	//Open BlockChain File and load it into a buffer
	reader := s.stream
	blocks := 0
	parsed := 0
	for {
		//log.Printf("Block #%v\n", blocks)
		/*if parsed > 1 {*/
		//break
		/*}*/
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
			if blocks < 100000 {
				continue
			}
			//Send the retrieved block to the parser
			s.wg.Add(1)
			go s.ParseBlock(blocks, block)
			parsed++
		} else {
			log.Printf("Invalid MagicID Looking for next block")
		}
	}

	return nil
}

//BlockData Structure
type Block struct {
	MagicID          [4]byte
	BlockLength      [4]uint8
	VersionNumber    [4]uint8
	PreviousHash     [32]uint8
	MerkleRoot       [32]uint8
	TimeStamp        [4]uint8
	TargetDifficulty [4]uint8
	Nonce            [4]uint8
	TransactionCount []uint8
	Transactions     []Transaction
}

func (b *Block) Hash() []byte {
	var d []uint8
	//d = append(d, b.magicid[:]...)
	//d = append(d, b.blocklength[:]...)
	d = append(d, b.VersionNumber[:]...)
	d = append(d, b.PreviousHash[:]...)
	d = append(d, b.MerkleRoot[:]...)
	d = append(d, b.TimeStamp[:]...)
	d = append(d, b.TargetDifficulty[:]...)
	d = append(d, b.Nonce[:]...)
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
func (b *Block) MagicIDVal() uint32 {
	var v uint32
	reader := bytes.NewReader(b.MagicID[:])
	if err := binary.Read(reader, binary.LittleEndian, &v); err != nil {
		log.Printf("MagicID Error: %v", err)
	}
	return v
}
func (b *Block) BlockLengthVal() uint32 {
	var v uint32
	reader := bytes.NewReader(b.BlockLength[:])
	if err := binary.Read(reader, binary.LittleEndian, &v); err != nil {
		log.Printf("BlockLength Error: %v", err)
	}
	return v
}
func (b *Block) VersionNumberVal() uint32 {
	var v uint32
	reader := bytes.NewReader(b.VersionNumber[:])
	if err := binary.Read(reader, binary.LittleEndian, &v); err != nil {
		log.Printf("VersionNumber Error: %v", err)
	}
	return v
}

func (b *Block) PreviousHashString() string {
	var temp []byte
	//Not sure how else to converte little endian to string.
	for i := 0; i < cap(b.PreviousHash); i++ {
		temp = append([]byte{b.PreviousHash[i]}, temp...)
	}
	return fmt.Sprintf("%x", temp[:])
}

func (b *Block) MerkleRootString() string {
	var temp []byte
	//Not sure how else to converte little endian to string.
	for i := 0; i < cap(b.MerkleRoot); i++ {
		temp = append([]byte{b.MerkleRoot[i]}, temp...)
	}
	return fmt.Sprintf("%x", temp[:])
}

func (b *Block) TimeStampVal() uint32 {
	var v uint32
	reader := bytes.NewReader(b.TimeStamp[:])
	if err := binary.Read(reader, binary.LittleEndian, &v); err != nil {
		log.Printf("TimeStamp Error: %v", err)
	}
	return v
}

func (b *Block) TimeStampFormatted() time.Time {
	if b.TimeStampVal() > 0 {
		return time.Unix(int64(b.TimeStampVal()), 0)
	} else {
		return time.Time{}
	}
}

func (b *Block) TargetDifficultyVal() uint32 {
	var v uint32
	reader := bytes.NewReader(b.TargetDifficulty[:])
	if err := binary.Read(reader, binary.LittleEndian, &v); err != nil {
		log.Printf("TargetDifficulty Error: %v", err)
	}
	return v

}

func (b *Block) NonceVal() uint32 {
	var v uint32
	reader := bytes.NewReader(b.Nonce[:])
	if err := binary.Read(reader, binary.LittleEndian, &v); err != nil {
		log.Printf("Nonce Error: %v", err)
	}
	return v
}

func (b *Block) TransactionCountVal() int {
	return VarInt(b.TransactionCount)
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

	if err := binary.Read(w, binary.LittleEndian, &block.MagicID); err != nil {
		return &block, err
	}

	if err := binary.Read(w, binary.LittleEndian, &block.BlockLength); err != nil {
		return &block, err
	}
	if err := binary.Read(w, binary.LittleEndian, &block.VersionNumber); err != nil {
		return &block, err
	}
	if err := binary.Read(w, binary.LittleEndian, &block.PreviousHash); err != nil {
		return &block, err
	}

	if err := binary.Read(w, binary.LittleEndian, &block.MerkleRoot); err != nil {
		return &block, err
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

	if b, err := w.ReadByte(); err != nil {
		log.Printf("Transaction Count Read Error: %v\n", err)
	} else {
		var c int
		block.TransactionCount = append(block.TransactionCount, b)
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
				block.TransactionCount = append(block.TransactionCount, d...)
			}
		}
	}

	for i := 0; i < block.TransactionCountVal(); i++ {
		if t, err := w.DecodeTrans(); err == nil {
			block.Transactions = append(block.Transactions, t)
		} else if err != io.EOF {
			log.Printf("Transaction Parse Error: %v\n", err)
		}
	}
	go w.PrintBlockInfo(block)
	return &block, nil
}

func (w *BlockParser) PrintBlockInfo(block Block) {
	blockOutputLog := []string{}

	blockOutputLog = append(blockOutputLog, fmt.Sprintf("Magic ID: %x\n", block.MagicIDVal()))
	blockOutputLog = append(blockOutputLog, fmt.Sprintf("Hash: %v\n", block.HashString()))
	blockOutputLog = append(blockOutputLog, fmt.Sprintf("Block Length: %v\n", block.BlockLengthVal()))
	blockOutputLog = append(blockOutputLog, fmt.Sprintf("Version Number: %v\n", block.VersionNumberVal()))
	blockOutputLog = append(blockOutputLog, fmt.Sprintf("Previous Hash: %v\n", block.PreviousHashString()))
	blockOutputLog = append(blockOutputLog, fmt.Sprintf("Merkle Root: %v\n", block.MerkleRootString()))
	blockOutputLog = append(blockOutputLog, fmt.Sprintf("TimeStamp: %v\n", block.TimeStampFormatted()))
	blockOutputLog = append(blockOutputLog, fmt.Sprintf("Target Difficulty: %v\n", block.TargetDifficultyVal()))
	blockOutputLog = append(blockOutputLog, fmt.Sprintf("Nonce: %v\n", block.NonceVal()))
	blockOutputLog = append(blockOutputLog, fmt.Sprintf("TransactionCount: %v\n", block.TransactionCountVal()))
	for i, t := range block.Transactions {
		blockOutputLog = append(
			blockOutputLog,
			fmt.Sprintf("Transaction %v:\n", i),
		)
		blockOutputLog = append(
			blockOutputLog,
			fmt.Sprintf("\tTransaction Version Number: %v\n", t.VersionNumber()),
		)
		blockOutputLog = append(
			blockOutputLog,
			fmt.Sprintf("\tInput Count: %v\n", t.InputCount()),
		)
		blockOutputLog = append(
			blockOutputLog,
			fmt.Sprintf("\tOutput Count: %v\n", t.OutputCount()),
		)
		blockOutputLog = append(
			blockOutputLog,
			fmt.Sprintf("\tTrsansaction Lock Time: %v\n", t.LockTimeFormatted()),
		)
		blockOutputLog = append(
			blockOutputLog,
			fmt.Sprintf("\tTransaction Hash: %v\n", t.HashString()),
		)
	}
	blockOutputLog = append(
		blockOutputLog,
		fmt.Sprint("####################"),
	)

	//fmt.Printf("%s\n\n", blockOutputLog)
}

func (w *BlockParser) DecodeTrans() (Transaction, error) {
	trans := Transaction{}

	if err := binary.Read(w, binary.LittleEndian, &trans.versionnumber); err != nil {
		log.Printf("Transaction Version Error: %v\n", err)
		return trans, err
	}

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
	return trans, nil
}

func (w *BlockParser) DecodeInput() (TransInput, error) {
	input := TransInput{}
	if err := binary.Read(w, binary.LittleEndian, &input.hash); err != nil {
		log.Printf("Input Hash Error: %v\n", err)
		return input, err
	}
	//log.Printf("\tInput Hash: %v\n", input.HashString())
	if err := binary.Read(w, binary.LittleEndian, &input.index); err != nil {
		log.Printf("Input Index Error: %v\n", err)
		return input, err
	}
	//log.Printf("\tInput Index: %v\n", input.Index())

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
		//log.Printf("\tScript Length: %v\n", input.ScriptLength())

		for i := 0; i < input.ScriptLength(); i++ {
			if b, err := w.ReadByte(); err == nil {
				input.script = append(input.script, b)
			}
		}
		//log.Printf("\tScript: %v\n", input.Script())
	}

	if err := binary.Read(w, binary.LittleEndian, &input.sequencenumber); err != nil {
		log.Printf("Input Index Error: %v\n", err)
		return input, err
	}
	//log.Printf("\tSequence Number: %v\n", input.SequenceNumber())

	return input, nil
}

func (w *BlockParser) DecodeOutput() (TransOutput, error) {
	out := TransOutput{}
	if err := binary.Read(w, binary.LittleEndian, &out.value); err != nil {
		log.Printf("Output Value Error: %v\n", err)
		return out, err
	}
	//log.Printf("\tOutput Value: %v\n", out.value)

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
	//log.Printf("\tScript Length: %v\n", VarInt(out.scriptlength))

	for i := 0; i < VarInt(out.scriptlength); i++ {
		if b, err := w.ReadByte(); err == nil {
			out.script = append(out.script, b)
		}
	}
	//log.Printf("\tScript: %v\n", out.script)

	return out, nil
}
