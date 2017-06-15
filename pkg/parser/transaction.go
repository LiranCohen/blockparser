package parser

import (
	"fmt"
	"bytes"
	"encoding/binary"
	"log"
	"crypto/sha256"
	"time"
	"github.com/lirancohen/blockparser/pkg/utils"
)

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
	return utils.VarInt(t.inputcount)
}

func (t *Transaction) OutputCount() int {
	return utils.VarInt(t.outputcount)
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
	return utils.VarInt(ti.scriptlength)
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
	return utils.VarInt(to.scriptlength)
}

func (to *TransOutput) Script() []uint8 {
	return to.script
}

