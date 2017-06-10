package parser

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"sync"
	"strings"
)


type BlockParser struct {
	*bufio.Reader
	wg *sync.WaitGroup
}

func NewBlockParser(r io.Reader, wg *sync.WaitGroup) *BlockParser {
	return &BlockParser{
		Reader: bufio.NewReader(r),
		wg: wg,
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
				if err == io.EOF {
					block.TransactionCount = make([]uint8, 0)
					return &block, nil
				}
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
		} else if err == io.EOF {
			log.Printf("Transaction EOF: %v\n", err)
			return &block, err;
		}
	}
	w.wg.Add(1)
	go w.PrintBlockInfo(block)
	return &block, nil
}

func (w *BlockParser) PrintBlockInfo(block Block) {
	defer w.wg.Done()
	blockOutputLog := []string{}

	blockOutputLog = append(
		blockOutputLog,
		fmt.Sprintf("\n####################START BLOCK####################\n"),
	)

	blockOutputLog = append(blockOutputLog, fmt.Sprintf("Magic ID: %x\n", block.MagicIDVal()))
	blockOutputLog = append(blockOutputLog, fmt.Sprintf("Hash: %v\n", block.HashString()))
	blockOutputLog = append(blockOutputLog, fmt.Sprintf("Block Length: %v\n", block.BlockLengthVal()))
	blockOutputLog = append(blockOutputLog, fmt.Sprintf("Version Number: %v\n", block.VersionNumberVal()))
	blockOutputLog = append(blockOutputLog, fmt.Sprintf("Previous Hash: %v\n", block.PreviousHashString()))
	blockOutputLog = append(blockOutputLog, fmt.Sprintf("Merkle Root: %v\n", block.MerkleRootString()))
	blockOutputLog = append(blockOutputLog, fmt.Sprintf("TimeStamp: %v\n", block.TimeStampFormatted()))
	blockOutputLog = append(blockOutputLog, fmt.Sprintf("Target Difficulty: %x\n", block.TargetDifficultyVal()))
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
		fmt.Sprintf("\n######################END BLOCK####################\n"),
	)

	log.Printf(strings.Join(blockOutputLog, ""))
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
			return trans, err
		} else {
			log.Printf("Transaction Parse Error: %v\n", err)
			return trans, err
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
