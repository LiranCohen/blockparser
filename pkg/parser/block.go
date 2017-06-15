package parser

import (
	"fmt"
	"bytes"
	"encoding/binary"
	"log"
	"time"
	"crypto/sha256"
	"github.com/lirancohen/blockparser/pkg/utils"
	"strings"
)

//BlockData Structure
type Block struct {
	Height		 int
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
	//Not sure how else to convert little endian to string.
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
	//Not sure how else to convert little endian to string.
	for i := 0; i < cap(b.PreviousHash); i++ {
		temp = append([]byte{b.PreviousHash[i]}, temp...)
	}
	return fmt.Sprintf("%x", temp[:])
}

func (b *Block) MerkleRootString() string {
	var temp []byte
	//Not sure how else to convert little endian to string.
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
	return utils.VarInt(b.TransactionCount)
}
func (b *Block) PrintBlockInfo() string {
	blockOutputLog := []string{}

	blockOutputLog = append(
		blockOutputLog,
		fmt.Sprintf("\n#################### START BLOCK %v ####################\n", b.Height),
	)

	blockOutputLog = append(blockOutputLog, fmt.Sprintf("Magic ID: %x\n", b.MagicIDVal()))
	blockOutputLog = append(blockOutputLog, fmt.Sprintf("Hash: %v\n", b.HashString()))
	blockOutputLog = append(blockOutputLog, fmt.Sprintf("Block Length: %v\n", b.BlockLengthVal()))
	blockOutputLog = append(blockOutputLog, fmt.Sprintf("Version Number: %v\n", b.VersionNumberVal()))
	blockOutputLog = append(blockOutputLog, fmt.Sprintf("Previous Hash: %v\n", b.PreviousHashString()))
	blockOutputLog = append(blockOutputLog, fmt.Sprintf("Merkle Root: %v\n", b.MerkleRootString()))
	blockOutputLog = append(blockOutputLog, fmt.Sprintf("TimeStamp: %v\n", b.TimeStampFormatted()))
	blockOutputLog = append(blockOutputLog, fmt.Sprintf("Target Difficulty: %x\n", b.TargetDifficultyVal()))
	blockOutputLog = append(blockOutputLog, fmt.Sprintf("Nonce: %v\n", b.NonceVal()))
	blockOutputLog = append(blockOutputLog, fmt.Sprintf("TransactionCount: %v\n", b.TransactionCountVal()))

	for i, t := range b.Transactions {
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
		fmt.Sprintf("\n#################### END BLOCK %v ####################\n", b.Height),
	)

	return strings.Join(blockOutputLog, "")
}
